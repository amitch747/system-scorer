package collector

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type cpuTimes struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

var (
	prevCPUTimes cpuTimes
	lastCPUExec  float64 // Cached value to avoid double scrape
)

type CPUCollector struct {
	cpuCountDesc *prometheus.Desc
	cpuExecDesc  *prometheus.Desc
}

func NewCPUCollector() *CPUCollector {
	return &CPUCollector{
		cpuCountDesc: prometheus.NewDesc(
			"syscore_cpu_count",
			"Number of CPU cores",
			nil,
			nil,
		),
		cpuExecDesc: prometheus.NewDesc(
			"syscore_cpu_exec",
			"15s percentage of CPU time spent not in idle or iowait (0-100)",
			nil,
			nil,
		),
	}
}

func (cc *CPUCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(cc, ch)
}

func (cc *CPUCollector) Collect(ch chan<- prometheus.Metric) {
	// Collect cpuCount
	ch <- prometheus.MustNewConstMetric(
		cc.cpuCountDesc,
		prometheus.GaugeValue,
		float64(runtime.NumCPU()),
	)

	currCPUTimes, err := readCPUTimes()
	if err != nil {
		return
	}
	cpuExec := calcCPUExec(prevCPUTimes, currCPUTimes)
	// Cache value for use in score.go
	lastCPUExec = cpuExec
	// Update for next scrape
	prevCPUTimes = currCPUTimes
	// Collect cpuExec as percentage (0-100) for Prometheus
	ch <- prometheus.MustNewConstMetric(
		cc.cpuExecDesc,
		prometheus.GaugeValue,
		cpuExec*100,
	)
}

func readCPUTimes() (cpuTimes, error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		return cpuTimes{}, err
	}
	defer file.Close()

	// Wrap file in scanner
	scanner := bufio.NewScanner(file)
	var cpuT cpuTimes
	// Read until there are no more lines
	for scanner.Scan() {
		// Seperate scanned line in a slice
		fields := strings.Fields(scanner.Text())
		if fields[0] == "cpu" {
			// Convert each uint64 into a string
			cpuT.user, _ = strconv.ParseUint(fields[1], 10, 64)
			cpuT.nice, _ = strconv.ParseUint(fields[2], 10, 64)
			cpuT.system, _ = strconv.ParseUint(fields[3], 10, 64)
			cpuT.idle, _ = strconv.ParseUint(fields[4], 10, 64)
			cpuT.iowait, _ = strconv.ParseUint(fields[5], 10, 64)
			cpuT.irq, _ = strconv.ParseUint(fields[6], 10, 64)
			cpuT.softirq, _ = strconv.ParseUint(fields[7], 10, 64)
			cpuT.steal, _ = strconv.ParseUint(fields[8], 10, 64)
			return cpuT, nil
		}
	}
	return cpuTimes{}, fmt.Errorf("cpu line not found")
}

func (cpuT cpuTimes) CalcTotalCPUTime() uint64 {
	return (cpuT.user + cpuT.nice + cpuT.system + cpuT.idle + cpuT.iowait + cpuT.irq + cpuT.softirq + cpuT.steal)
}

func calcCPUExec(prev, curr cpuTimes) float64 {
	// First need total time passed
	totalPrev := prev.CalcTotalCPUTime()
	totalCurr := curr.CalcTotalCPUTime()

	totalDelta := float64(totalCurr - totalPrev)
	if totalDelta <= 0 {
		return 0
	}
	// idleTime is idle AND iowait
	idleTime := float64((curr.idle - prev.idle) + (curr.iowait - prev.iowait))
	// Return as 0-1 for internal use
	return 1 - idleTime/totalDelta
}
