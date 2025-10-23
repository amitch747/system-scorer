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

// Need to calc total CPU time not spent idle or waiting during our previous scrape interval (15s?)

type cpuTimes struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

var prevCPUTimes cpuTimes

type CPUCollector struct {
	cpuCountDesc          *prometheus.Desc
	cpuExecPercentageDesc *prometheus.Desc
}

func NewCPUCollector() *CPUCollector {
	return &CPUCollector{
		cpuCountDesc: prometheus.NewDesc(
			"syscraper_cpu_count",
			"Number of CPU cores",
			nil,
			nil,
		),
		cpuExecPercentageDesc: prometheus.NewDesc(
			"syscraper_cpu_exec_percentage",
			"15s percentage of CPU time spent not in idle or iowait",
			nil,
			nil,
		),
	}
}

func (cc CPUCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(cc, ch)
}

func (cc CPUCollector) Collect(ch chan<- prometheus.Metric) {
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
	cpuExecPercentage := calcCPUExecPercentage(prevCPUTimes, currCPUTimes)
	// Update for next scrape
	prevCPUTimes = currCPUTimes
	// Collect cpuExecPercentage
	ch <- prometheus.MustNewConstMetric(
		cc.cpuExecPercentageDesc,
		prometheus.GaugeValue,
		float64(cpuExecPercentage),
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
	// Read until there are no more lines
	for scanner.Scan() {
		// Seperate scanned line in a slice
		fields := strings.Fields(scanner.Text())
		if fields[0] == "cpu" {
			var cpuT cpuTimes
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

func (cpuT cpuTimes) calcTotalCPUTime() uint64 {
	return (cpuT.user + cpuT.nice + cpuT.system + cpuT.idle + cpuT.iowait + cpuT.irq + cpuT.softirq + cpuT.steal)
}

func calcCPUExecPercentage(prev, curr cpuTimes) float64 {
	// First need total time passed
	totalPrev := prev.calcTotalCPUTime()
	totalCurr := curr.calcTotalCPUTime()

	totalDelta := float64(totalCurr - totalPrev)
	if totalDelta <= 0 {
		return 0
	}
	// idleTime is idle AND iowait
	idleTime := float64((curr.idle - prev.idle) + (curr.iowait - prev.iowait))
	return 100 * (1 - idleTime/totalDelta)
}
