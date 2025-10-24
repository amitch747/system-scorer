package collector

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/amitch747/prometheus-system-scraper/utility"
	"github.com/prometheus/client_golang/prometheus"
)

type diskStats struct {
	name         string
	ioTime       uint64
	weightedTime uint64
}

type ioCollector struct {
	maxIoTimeDesc     *prometheus.Desc
	maxQueueDepthDesc *prometheus.Desc
}

func NewIoCollector() *ioCollector {
	return &ioCollector{
		maxIoTimeDesc: prometheus.NewDesc(
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
	disks, err := readDiskstats()
	if err != nil {
		return
	}

	// process disks

	ch <- prometheus.MustNewConstMetric(
		ic.ioTimeDesc,
		prometheus.GaugeValue,
		ioTime,
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

// keep a slice of previous diskStats
var prevDiskStats []diskStats
var scrape_interval = 15

// Looking for bottlenecks
func calcDisk(prev, curr []diskStats) (float64, float64) {
	var deltaIoSlice []float64
	var deltaWeightedSlice []float64

	// To avoid nested for loops, store previous disk states in a map
	prevMap := make(map[string]diskStats)
	for _, disk := range prev {
		prevMap[disk.name] = disk
	}

	for _, currDisk := range curr {
		prevDisk, exists := prevMap[currDisk.name]
		if !exists {
			continue // Must be a new disk
		}

		deltaIoTime := currDisk.ioTime - prevDisk.ioTime
		ioUtilPercent := (float64(deltaIoTime) / (float64(scrape_interval)) * 100)
		if ioUtilPercent > 100.0 {
			ioUtilPercent = 100.0 // Need to clamp to avoid jitter
		}
		deltaIoSlice = append(deltaIoSlice, float64(deltaIoTime))

		blockDeviceInfo := utility.GetBlockDeviceInfoMap()

		deltaWeightedTime := currDisk.weightedTime - prevDisk.weightedTime
		deltaWeightedSlice = append(deltaWeightedSlice, float64(deltaWeightedTime))
		avgQueueDepth := float64(deltaWeightedTime) / float64(deltaIoTime) // gives number of concurrent IO
		k := 5.0

		// Store current times globally
		prevDiskStats = append(prevDiskStats, diskStats{
			name:         currDisk.name,
			ioTime:       currDisk.ioTime,
			weightedTime: currDisk.weightedTime,
		})
	}
	// max queue depth
	maxIoTime = max
	maxQueueDepthDesc
	// return
}
