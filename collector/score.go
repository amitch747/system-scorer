package collector

import (
	"math"
	"os"
	"runtime"

	"github.com/amitch747/system-scorer/utility"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs/sysfs"
)

/*
1) syscore_utilization_score_weighted
   - Smooth nonlinear weighted average
   - Represents "overall load"

2) syscore_utilization_score_bottleneck
   - Soft-OR / bottleneck emphasis
   - Represents "effective bottleneck pressure"

Both are 0–100 scores consumed by dashboards.
*/

type scoreCollector struct {
	weightedScoreDesc   *prometheus.Desc
	bottleneckScoreDesc *prometheus.Desc
}

func NewScoreCollector() *scoreCollector {
	return &scoreCollector{
		weightedScoreDesc: prometheus.NewDesc(
			"syscore_utilization_score_weighted",
			"Nonlinear weighted utilization score (0–100)",
			nil, nil,
		),
		bottleneckScoreDesc: prometheus.NewDesc(
			"syscore_utilization_score_bottleneck",
			"Bottleneck-or utilization score (0–100)",
			nil, nil,
		),
	}
}

func (sc *scoreCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sc.weightedScoreDesc
	ch <- sc.bottleneckScoreDesc
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
	return calcCPUExecPercentage(prevCPUTimes, curr)
}

func getMemoryUtilization() float64 {
	m, err := readMemInfo()
	if err != nil || m.memTotal == 0 {
		return 0
	}
	return float64(m.memTotal-m.memAvailable) / float64(m.memTotal) * 100
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
			vram = float64(card.MemoryVRAMUsed) / float64(card.MemoryVRAMSize) * 100
		}
		// 70% busy, 30% VRAM - maybe change
		sum += 0.7*float64(card.GPUBusyPercent) + 0.3*vram
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
	entires, err := os.ReadDir("/run/user")
	if err != nil {
		return 0
	}

	userCount := len(entires)
	var userUtil float64
	// GPU Node
	gpuNode, gpuCount := utility.GetGPUConfig()
	if gpuNode {
		// 1 card per person
		userUtil = float64(userCount) / float64(gpuCount) * 100
	} else {
		cpuCapacity := runtime.NumCPU() / 16
		userUtil = float64(userCount) / float64(cpuCapacity) * 100
	}
	// CPU Node
	// 16 cores per user

	if userUtil > 100 {
		userUtil = 100
	}
	return userUtil

}

func calcWeightedScore(cpu, mem, gpu, io, net, user float64, hasGPU bool) float64 {

	// Nonlinear (higher util penalized more)
	c := math.Pow(cpu/100, 1.2)
	m := math.Pow(mem/100, 1.5)
	d := math.Pow(io/100, 1.2)
	// Exponential saturation to reflect network congestion
	n := 1 - math.Exp(-2*(net/100))
	g := 0.0
	if hasGPU {
		g = math.Pow(gpu/100, 1.2)
	}
	u := 1 - math.Exp(-2*(user/100))

	// Weights emphasize GPU > CPU > Mem > IO > Net
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

	c := math.Pow(cpu/100, 1.2)
	m := math.Pow(mem/100, 1.5)
	d := math.Pow(io/100, 1.2)
	n := 1 - math.Exp(-2*(net/100))
	g := 0.0
	if hasGPU {
		g = math.Pow(gpu/100, 1.2)
	}
	u := 1 - math.Exp(-2*(user/100))

	// Soft-OR = bottleneck emphasis
	bottleneck := math.Max(c, math.Max(m, math.Max(d, math.Max(n, math.Max(g, u)))))

	return bottleneck * 100
}
