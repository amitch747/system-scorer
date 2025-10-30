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
	memUsageDesc    *prometheus.Desc
	memCommitDesc   *prometheus.Desc
	memSwapUsedDesc *prometheus.Desc
	memPressureDesc *prometheus.Desc
}

func NewMemCollector() *memCollector {
	return &memCollector{
		memUsageDesc: prometheus.NewDesc(
			"syscore_mem_usage",
			"Percentage of physical memory in use",
			nil,
			nil,
		),
		memCommitDesc: prometheus.NewDesc(
			"syscore_mem_commit",
			"Percentage of committed virtual memory over commit limit",
			nil,
			nil,
		),
		memSwapUsedDesc: prometheus.NewDesc(
			"syscore_mem_swap",
			"Percentage of swap space in use",
			nil,
			nil,
		),
		memPressureDesc: prometheus.NewDesc(
			"syscore_mem_pressure",
			"Weighted memory pressure index (usage + swap + commit)",
			nil,
			nil,
		),
	}
}

func (mc *memCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(mc, ch)
}

func (mc *memCollector) Collect(ch chan<- prometheus.Metric) {
	mInfo, err := readMemInfo()
	if err != nil {
		return
	}
	// Collect memUsage
	var memUsed float64
	// Make sure denominators not zero
	if mInfo.memTotal > 0 {
		memUsed = float64(mInfo.memTotal-mInfo.memAvailable) / float64(mInfo.memTotal)
	}
	ch <- prometheus.MustNewConstMetric(
		mc.memUsageDesc,
		prometheus.GaugeValue,
		memUsed*100,
	)
	// Collect commitRatio
	var memCommit float64
	if mInfo.commitLimit > 0 {
		memCommit = float64(mInfo.commitAS) / float64(mInfo.commitLimit)
	}
	ch <- prometheus.MustNewConstMetric(
		mc.memCommitDesc,
		prometheus.GaugeValue,
		memCommit*100,
	)
	// Collect swapRatio
	var memSwap float64
	if mInfo.swapTotal > 0 {
		memSwap = float64(mInfo.swapTotal-mInfo.swapFree) / float64(mInfo.swapTotal)
	}
	ch <- prometheus.MustNewConstMetric(
		mc.memSwapUsedDesc,
		prometheus.GaugeValue,
		memSwap*100,
	)

	// nonlinear scaling (high commitRatio is very bad)
	scaledMem := math.Pow(memUsed, 1.5)
	scaledCommit := math.Pow(memCommit, 2.5)
	scaledSwap := math.Pow(memSwap, 2.0)
	// Saturating exponential. Needs tweaking
	memPressure := 1 - math.Exp(-3*(0.7*scaledMem+0.2*scaledCommit+0.1*scaledSwap))

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
