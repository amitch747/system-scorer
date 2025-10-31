package collector

import (
	"math"
	"runtime"

	"github.com/amitch747/system-scorer/utility"
	"github.com/prometheus/client_golang/prometheus"
)

type scoreCollector struct {
	weightedScoreDesc *prometheus.Desc
	cpuUtilDesc       *prometheus.Desc
	gpuUtilDesc       *prometheus.Desc
	memUtilDesc       *prometheus.Desc
	ioUtilDesc        *prometheus.Desc
	netUtilDesc       *prometheus.Desc
	userUtilDesc      *prometheus.Desc
}

func NewScoreCollector() *scoreCollector {
	return &scoreCollector{
		weightedScoreDesc: prometheus.NewDesc(
			"syscore_utilization_score_weighted",
			"Nonlinear weighted utilization score (0–100)",
			nil,
			nil,
		),
		cpuUtilDesc: prometheus.NewDesc(
			"syscore_scaled_cpu_util",
			"Scaled CPU exec time ratio used in utilization score",
			nil,
			nil,
		),
		gpuUtilDesc: prometheus.NewDesc(
			"syscore_scaled_gpu_util",
			"Scaled average of GPU util (busy % and VRAM) used in utilization score",
			nil,
			nil,
		),
		memUtilDesc: prometheus.NewDesc(
			"syscore_scaled_mem_util",
			"Scaled memory usage ratio used in utilization score",
			nil,
			nil,
		),
		ioUtilDesc: prometheus.NewDesc(
			"syscore_scaled_io_util",
			"Scaled max IO util (see io.go) used in utilization score",
			nil,
			nil,
		),
		netUtilDesc: prometheus.NewDesc(
			"syscore_scaled_net_util",
			"Scaled CPU exec time used in utilization score",
			nil,
			nil,
		),
		userUtilDesc: prometheus.NewDesc(
			"syscore_user_util",
			"Ratio of user count to available hardware (1 GPU/user or 16 CPU/user)",
			nil,
			nil,
		),
	}
}

func (sc *scoreCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(sc, ch)
}

func (sc *scoreCollector) Collect(ch chan<- prometheus.Metric) {

	// Gather info from other collectors
	gpuUtil := SharedGpuUtil
	hasGPU := GPUCheck
	cpuUtil := SharedCPUExec
	memUtil := SharedMemUsed
	ioUtil := SharedMaxIOTime
	netUtil := SharedMaxNetSaturation

	// Calculate user util
	userUtil := getUserUtilization()

	// Scale utilization values
	scaledUtils := utilScaling(gpuUtil, cpuUtil, memUtil, ioUtil, netUtil, hasGPU)

	// Export scaled GPU
	ch <- prometheus.MustNewConstMetric(
		sc.gpuUtilDesc, prometheus.GaugeValue, scaledUtils.g,
	)
	// Export scaled CPU
	ch <- prometheus.MustNewConstMetric(
		sc.cpuUtilDesc, prometheus.GaugeValue, scaledUtils.c,
	)
	// Export scaled memory
	ch <- prometheus.MustNewConstMetric(
		sc.memUtilDesc, prometheus.GaugeValue, scaledUtils.m,
	)
	// Export scaled I/O
	ch <- prometheus.MustNewConstMetric(
		sc.ioUtilDesc, prometheus.GaugeValue, scaledUtils.i,
	)
	// Export scaled network
	ch <- prometheus.MustNewConstMetric(
		sc.netUtilDesc, prometheus.GaugeValue, scaledUtils.n,
	)
	// Export user count
	ch <- prometheus.MustNewConstMetric(
		sc.userUtilDesc, prometheus.GaugeValue, userUtil,
	)

	// Calcualte weighted utilization score
	weighted := calcWeightedScore(scaledUtils, userUtil, hasGPU)

	ch <- prometheus.MustNewConstMetric(
		sc.weightedScoreDesc, prometheus.GaugeValue, weighted,
	)

}

// func getGPUUtilization() (float64, bool) {
// 	fs, err := sysfs.NewFS("/sys")
// 	if err != nil {
// 		return 0, false
// 	}
// 	stats, err := fs.ClassDRMCardAMDGPUStats()
// 	if err != nil || len(stats) == 0 {
// 		return 0, false
// 	}

// 	var sum float64
// 	for _, card := range stats {
// 		var vram float64
// 		if card.MemoryVRAMSize > 0 {
// 			vram = float64(card.MemoryVRAMUsed) / float64(card.MemoryVRAMSize)
// 		}
// 		// 70% busy, 30% VRAM - maybe change
// 		// GPUBusyPercent is already 0-100, normalize to 0-1
// 		sum += 0.7*float64(card.GPUBusyPercent)/100 + 0.3*vram
// 	}

// 	return sum / float64(len(stats)), true
// }

// func getDiskUtilization() float64 {
// 	curr, err := readDiskstats()
// 	if err != nil {
// 		return 0
// 	}
// 	maxIO, _ := calcDisk(prevDiskStats, curr)
// 	return maxIO
// }

// func getNetworkUtilization() float64 {
// 	stats, err := readNetworkStats()
// 	if err != nil {
// 		return 0
// 	}

// 	speeds := utility.GetLinkSpeeds()
// 	devs := calcNetworkMetrics(stats, speeds)

// 	max := 0.0
// 	for _, m := range devs {
// 		if m.saturationPercentage > max {
// 			max = m.saturationPercentage
// 		}
// 	}
// 	return max
// }

func getUserUtilization() float64 {
	userCount := SharedUserCount

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

type scaledUtilizations struct {
	g, c, m, i, n float64
}

func calcWeightedScore(sacledUtils scaledUtilizations, usersUtil float64, hasGPU bool) float64 {

	// Setup weights
	// emphasize GPU > CPU > Mem > IO >= Net
	var wGPU, wCPU, wMem, wIO, wNet, wUser float64
	if hasGPU {
		wGPU, wCPU, wMem, wIO, wNet, wUser = 0.34, 0.20, 0.10, 0.01, 0.01, 0.34
	} else {
		wGPU, wCPU, wMem, wIO, wNet, wUser = 0.0, 0.54, 0.10, 0.01, 0.01, 0.34
	}

	// Soft aggregation (smooth AND)
	score := 1 - ((1 - wCPU*sacledUtils.c) *
		(1 - wMem*sacledUtils.m) *
		(1 - wGPU*sacledUtils.g) *
		(1 - wIO*sacledUtils.i) *
		(1 - wNet*sacledUtils.n) *
		(1 - wUser*usersUtil))

	return score * 100
}

func utilScaling(cpu, mem, gpu, io, net float64, hasGPU bool) scaledUtilizations {

	// Nonlinear (higher util penalized more)
	scaledCPU := math.Pow(cpu, 1.2)
	scaledMem := math.Pow(mem, 1.5)
	scaledIO := math.Pow(io, 1.2)
	scaledNet := 1 - math.Exp(-2*net) // Exponential saturation for network congestion
	scaledGPU := 0.0
	if hasGPU {
		scaledGPU = math.Pow(gpu, 1.2)
	}

	return scaledUtilizations{
		g: scaledGPU,
		c: scaledCPU,
		m: scaledMem,
		i: scaledIO,
		n: scaledNet,
	}
}
