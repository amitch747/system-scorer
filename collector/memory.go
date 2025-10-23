package collector

import (
	"bufio"
	"os"
	"strconv"

	"math"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

/*
	from /proc/meminfo

	mem_score = 0.8*(MemTotal - MemAvailable)/MemTotal + 0.2*(1 - SwapFree/SwapTotal)

	commit_ratio = Committed_AS / CommitLimit

*/

type memInfo struct {
	memTotal, memAvailable, swapTotal, swapFree, commitLimit, commitAS uint64
}

type memCollector struct {
	memUsageDesc       *prometheus.Desc
	memCommitRatioDesc *prometheus.Desc
	memSwapUsedDesc    *prometheus.Desc
	memPressureDesc    *prometheus.Desc
}

func NewMemCollector() *memCollector {
	return &memCollector{
		memUsageDesc: prometheus.NewDesc(
			"syscraper_mem_usage",
			"Current ratio of physical memory in use",
			nil,
			nil,
		),
		memCommitRatioDesc: prometheus.NewDesc(
			"syscraper_mem_commit_ratio",
			"Current ratio of committed virtual memory to commit limit",
			nil,
			nil,
		),
		memSwapUsedDesc: prometheus.NewDesc(
			"syscraper_mem_swap_ratio",
			"Current ratio of swap space in use",
			nil,
			nil,
		),
		memPressureDesc: prometheus.NewDesc(
			"syscraper_mem_pressure",
			"Weighted memory pressure index (usage + swap + commit)",
			nil,
			nil,
		),
	}
}

func (mc memCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc memCollector) Collect(ch chan<- prometheus.Metric) {
	mInfo, err := readMemInfo()
	if err != nil {
		return
	}
	// Collect memUsage
	var usedRatio float64
	// Make sure denominators not zero
	if mInfo.memTotal > 0 {
		usedRatio = float64(mInfo.memTotal-mInfo.memAvailable) / float64(mInfo.memTotal)
	}
	ch <- prometheus.MustNewConstMetric(
		mc.memUsageDesc,
		prometheus.GaugeValue,
		usedRatio,
	)
	// Collect commitRatio
	var commitRatio float64
	if mInfo.commitLimit > 0 {
		commitRatio = float64(mInfo.commitAS) / float64(mInfo.commitLimit)
	}
	ch <- prometheus.MustNewConstMetric(
		mc.memCommitRatioDesc,
		prometheus.GaugeValue,
		commitRatio,
	)
	// Collect swapRatio
	var swapRatio float64
	if mInfo.swapTotal > 0 {
		swapRatio = float64(mInfo.swapTotal-mInfo.swapFree) / float64(mInfo.swapTotal)
	}
	ch <- prometheus.MustNewConstMetric(
		mc.memSwapUsedDesc,
		prometheus.GaugeValue,
		swapRatio,
	)

	// nonlinear scaling (high commitRatio is very bad)
	scaledMem := math.Pow(usedRatio, 1.5)
	scaledSwap := math.Pow(swapRatio, 2.0)
	scaledCommit := math.Pow(commitRatio, 3.0)
	memPressure := 0.7*scaledMem + 0.2*scaledCommit + 0.1*scaledSwap

	ch <- prometheus.MustNewConstMetric(
		mc.memPressureDesc,
		prometheus.GaugeValue,
		float64(memPressure),
	)
}

func readMemInfo() (memInfo, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return memInfo{}, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var m memInfo

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		val, _ := strconv.ParseUint(fields[1], 10, 64)
		switch fields[0] {
		case "MemTotal:":
			m.memTotal = val
		case "MemAvailable:":
			m.memAvailable = val
		case "SwapTotal:":
			m.swapTotal = val
		case "SwapFree:":
			m.swapFree = val
		case "CommitLimit:":
			m.commitLimit = val
		case "Committed_AS:":
			m.commitAS = val
		}
	}
	return m, scanner.Err()
}
