package collector

import (
	"bufio"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/amitch747/prometheus-system-scraper/utility"
	"github.com/prometheus/client_golang/prometheus"
)

type diskStats struct {
	name         string
	ioTime       uint64
	weightedTime uint64
}

// keep a slice of previous diskStats
var prevDiskStats []diskStats
var scrape_interval = 15
var diskInfoOnce sync.Once

type ioCollector struct {
	maxIOTimeDesc     *prometheus.Desc
	maxIOPressureDesc *prometheus.Desc
}

func NewIoCollector() *ioCollector {
	return &ioCollector{
		maxIOTimeDesc: prometheus.NewDesc(
			"syscraper_io_time",
			"15s interval of time spent doing IO",
			nil,
			nil,
		),
	}
}

func (ic ioCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(ic, ch)
}

func (ic ioCollector) Collect(ch chan<- prometheus.Metric) {
	currDiskStats, err := readDiskstats()
	if err != nil {
		return
	}

	// process disks
	maxIoTime, maxIOPressure := calcDisk(prevDiskStats, currDiskStats)

	ch <- prometheus.MustNewConstMetric(
		ic.maxIOTimeDesc,
		prometheus.GaugeValue,
		maxIoTime,
	)
}

func readDiskstats() ([]diskStats, error) {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		return []diskStats{}, err
	}
	defer file.Close()

	var disks []diskStats
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue // malformed/old kernel line
		}

		name := fields[2]
		if strings.HasPrefix(name, "ram") ||
			strings.HasPrefix(name, "loop") ||
			strings.HasPrefix(name, "/") ||
			// nasty
			len(name) > 3 && (name[0] == 's' && name[3] >= '0' && name[3] <= '9') {
			continue
		}

		ioTime, _ := strconv.ParseUint(fields[12], 10, 64)
		weightedTime, _ := strconv.ParseUint(fields[13], 10, 64)

		disks = append(disks, diskStats{
			name:         name,
			ioTime:       ioTime,
			weightedTime: weightedTime,
		})
	}
	return disks, scanner.Err()
}

// Looking for bottlenecks
func calcDisk(prev, curr []diskStats) (float64, float64) {
	var maxIoUtil, maxPressure float64
	blockDeviceInfoMap := utility.GetBlockDeviceInfoMap()

	// To avoid nested for loops, store previous disk states in a map
	prevMap := make(map[string]diskStats)
	for _, disk := range prev {
		prevMap[disk.name] = disk
	}

	for _, currDisk := range curr {
		prevDisk, exists := prevMap[currDisk.name]
		if !exists {
			prevDiskStats = append(prevDiskStats, currDisk)
			continue // Must be a new disk
		}

		deltaIoTime := currDisk.ioTime - prevDisk.ioTime
		ioUtil := (float64(deltaIoTime) / (float64(scrape_interval)) * 100)
		if ioUtil > 100.0 {
			ioUtil = 100.0 // Need to clamp to avoid jitter
		}
		if ioUtil > maxIoUtil {
			maxIoUtil = ioUtil
		}

		deltaWeightedTime := currDisk.weightedTime - prevDisk.weightedTime
		avgQueueDepth := float64(deltaWeightedTime) / float64(deltaIoTime) // gives number of concurrent IO

		// Check if name is rotational
		var pressure float64
		if base, ok := utility.MatchBaseDevice(currDisk.name, blockDeviceInfoMap); ok {
			blockDevice := blockDeviceInfoMap[base]
			pressure = ioPressure(avgQueueDepth, blockDevice.Type)
		}

		if pressure > maxPressure {
			maxPressure = pressure
		}

		// Store current times globally
		prevDiskStats = append(prevDiskStats, diskStats{
			name:         currDisk.name,
			ioTime:       currDisk.ioTime,
			weightedTime: currDisk.weightedTime,
		})
	}

	return maxIoUtil, maxPressure
}

func ioPressure(avgQueueDepth float64, deviceType string) float64 {
	var k float64
	switch deviceType {
	case "HDD":
		k = 2.0
	case "SDD":
		k = 5.0
	case "NVMe":
		k = 10.0
	default:
		k = 5.0
	}

	pressure := 1 - math.Exp(-avgQueueDepth/k)
	if pressure > 1 {
		pressure = 1
	}

	return pressure
}
