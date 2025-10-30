package collector

import (
	"math"
	"runtime"

	"github.com/amitch747/system-scorer/utility"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs/sysfs"
)

type scoreCollector struct {
	weightedScoreDesc   *prometheus.Desc
	bottleneckScoreDesc *prometheus.Desc
}

func NewScoreCollector() *scoreCollector {
	return &scoreCollector{
		weightedScoreDesc: prometheus.NewDesc(
			"syscore_utilization_score_weighted",
			"Nonlinear weighted utilization score (0–100)",
			nil,
			nil,
		),
		bottleneckScoreDesc: prometheus.NewDesc(
			"syscore_utilization_score_bottleneck",
			"Bottleneck utilization score (0–100)",
			nil,
			nil,
		),
	}
}

func (sc *scoreCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(sc, ch)
}

func (sc *scoreCollector) Collect(ch chan<- prometheus.Metric) {

	cpu := getCPUUtilization()
	mem := getMemoryUtilization()
	gpu, hasGPU := getGPUUtilization()
	io := getDiskUtilization()
	net := getNetworkUtilization()
	user := getUserUtilization()

	weighted := calcWeightedScore(cpu, mem, gpu, io, net, user, hasGPU)
	bottleneck := calcBottleneckScore(cpu, mem, gpu, io, net, user, hasGPU)

	ch <- prometheus.MustNewConstMetric(
		sc.weightedScoreDesc, prometheus.GaugeValue, weighted,
	)

	ch <- prometheus.MustNewConstMetric(
		sc.bottleneckScoreDesc, prometheus.GaugeValue, bottleneck,
	)
}

func getCPUUtilization() float64 {
	curr, err := readCPUTimes()
	if err != nil {
		return 0
	}
	return calcCPUExec(prevCPUTimes, curr)
}

func getMemoryUtilization() float64 {
	m, err := readMemInfo()
	if err != nil || m.memTotal == 0 {
		return 0
	}
	return float64(m.memTotal-m.memAvailable) / float64(m.memTotal)
}

func getGPUUtilization() (float64, bool) {
	fs, err := sysfs.NewFS("/sys")
	if err != nil {
		return 0, false
	}
	stats, err := fs.ClassDRMCardAMDGPUStats()
	if err != nil || len(stats) == 0 {
		return 0, false
	}

	var sum float64
	for _, card := range stats {
		var vram float64
		if card.MemoryVRAMSize > 0 {
			vram = float64(card.MemoryVRAMUsed) / float64(card.MemoryVRAMSize)
		}
		// 70% busy, 30% VRAM - maybe change
		// GPUBusyPercent is already 0-100, normalize to 0-1
		sum += 0.7*float64(card.GPUBusyPercent)/100 + 0.3*vram
	}

	return sum / float64(len(stats)), true
}

func getDiskUtilization() float64 {
	curr, err := readDiskstats()
	if err != nil {
		return 0
	}
	maxIO, _ := calcDisk(prevDiskStats, curr)
	return maxIO
}

func getNetworkUtilization() float64 {
	stats, err := readNetworkStats()
	if err != nil {
		return 0
	}

	speeds := utility.GetLinkSpeeds()
	devs := calcNetworkMetrics(stats, speeds)

	max := 0.0
	for _, m := range devs {
		if m.saturationPercentage > max {
			max = m.saturationPercentage
		}
	}
	return max
}

func getUserUtilization() float64 {
	userCount := GetActiveUserCount()

	gpuNode, gpuCount := utility.GetGPUConfig()

	var capacity int
	if gpuNode {
		// GPU Node: 1 user per GPU
		capacity = gpuCount
	} else {
		// CPU Node: 16 cores per user
		capacity = runtime.NumCPU() / 16
	}

	// Prevent division by zero
	if capacity == 0 {
		capacity = 1
	}

	userUtil := float64(userCount) / float64(capacity)

	// Clamp to 0-1 range
	if userUtil > 1.0 {
		userUtil = 1.0
	}

	return userUtil
}

func calcWeightedScore(cpu, mem, gpu, io, net, user float64, hasGPU bool) float64 {

	// Nonlinear (higher util penalized more)
	c := math.Pow(cpu, 1.2)
	m := math.Pow(mem, 1.5)
	d := math.Pow(io, 1.2)
	n := 1 - math.Exp(-2*net) // Exponential saturation for network congestion
	g := 0.0
	if hasGPU {
		g = math.Pow(gpu, 1.2)
	}
	u := user // Exponential saturation for user sessions

	// emphasize GPU > CPU > Mem > IO > Net
	var wCPU, wMem, wGPU, wDisk, wNet, wUser float64
	if hasGPU {
		wGPU, wCPU, wMem, wDisk, wNet, wUser = 0.34, 0.20, 0.10, 0.01, 0.01, 0.34
	} else {
		wGPU, wCPU, wMem, wDisk, wNet, wUser = 0.0, 0.54, 0.10, 0.01, 0.01, 0.34
	}

	// Soft aggregation (smooth AND)
	score := 1 -
		((1 - wCPU*c) *
			(1 - wMem*m) *
			(1 - wGPU*g) *
			(1 - wDisk*d) *
			(1 - wNet*n) *
			(1 - wUser*u))

	return score * 100
}

func calcBottleneckScore(cpu, mem, gpu, io, net, user float64, hasGPU bool) float64 {

	c := math.Pow(cpu, 1.2)
	m := math.Pow(mem, 1.5)
	d := math.Pow(io, 1.2)
	n := 1 - math.Exp(-2*net)
	g := 0.0
	if hasGPU {
		g = math.Pow(gpu, 1.2)
	}
	u := user

	// Soft-OR = bottleneck emphasis
	bottleneck := math.Max(c, math.Max(m, math.Max(d, math.Max(n, math.Max(g, u)))))

	return bottleneck * 100
}
