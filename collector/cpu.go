package collector

import "github.com/prometheus/client_golang/prometheus"

type CPUCollector struct {
	cpuCountDesc *prometheus.Desc
}

// fs, err := procfs.NewDefaultFS()
// stat, err := fs.Stat()

// stat.CPUtotal

func NewCPUCollect() *CPUCollector {
	return &CPUCollector{
		cpuCountDesc: prometheus.NewDesc(
			"cpu_count",
			"Number of CPU cores.",
			nil,
			nil,
		),
	}

}
