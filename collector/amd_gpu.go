package collector

import (
	"github.com/prometheus/client_golang/prometheus"

	"github.com/prometheus/procfs/sysfs"
)

type AMDGPUCollector struct {
	gpuBusyPercentDesc        *prometheus.Desc
	gpuGTTSizeDesc            *prometheus.Desc
	gpuGTTUsedDesc            *prometheus.Desc
	gpuVisibleVRAMSizeDesc    *prometheus.Desc
	gpuVisibleVRAMUsedDesc    *prometheus.Desc
	gpuMemoryVRAMSizeDesc     *prometheus.Desc
	gpuMemoryVRAMUsedDesc     *prometheus.Desc
	gpuAverageUtilizationDesc *prometheus.Desc
}

// Used in score.go to avoid double scrape
var SharedGpuUtil float64

// Potential future upgrades
// /sys/class/drm/card*/device/mem_busy_percent
// /sys/kernel/kfd

func NewAMDGPUCollector() (*AMDGPUCollector, error) {
	return &AMDGPUCollector{
		gpuBusyPercentDesc: prometheus.NewDesc(
			"syscore_gpu_busy_percent",
			"Percentage GPU is busy.",
			[]string{"card", "id"},
			nil,
		),
		gpuGTTSizeDesc: prometheus.NewDesc(
			"syscore_gpu_gtt_size",
			"Size of GTT block in bytes.",
			[]string{"card", "id"},
			nil,
		),
		gpuGTTUsedDesc: prometheus.NewDesc(
			"syscore_gpu_gtt_used",
			"Used bytes of GTT block.",
			[]string{"card", "id"},
			nil,
		),
		gpuVisibleVRAMSizeDesc: prometheus.NewDesc(
			"syscore_gpu_visible_vram_size",
			"Size of visible VRAM in bytes.",
			[]string{"card", "id"},
			nil,
		),
		gpuVisibleVRAMUsedDesc: prometheus.NewDesc(
			"syscore_gpu_visible_vram_used",
			"Used bytes of visible VRAM.",
			[]string{"card", "id"},
			nil,
		),
		gpuMemoryVRAMSizeDesc: prometheus.NewDesc(
			"syscore_gpu_vram_size",
			"Size of VRAM in bytes.",
			[]string{"card", "id"},
			nil,
		),
		gpuMemoryVRAMUsedDesc: prometheus.NewDesc(
			"syscore_gpu_vram_used",
			"Used bytes of VRAM.",
			[]string{"card", "id"},
			nil,
		),
		gpuAverageUtilizationDesc: prometheus.NewDesc(
			"syscore_gpu_avg_util",
			"Average percentage of gpu utilization (0-100)",
			nil,
			nil,
		),
	}, nil
}

func (gc *AMDGPUCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(gc, ch)
}

func (gc *AMDGPUCollector) Collect(ch chan<- prometheus.Metric) {
	fs, err := sysfs.NewFS("/sys")
	if err != nil {
		return
	}
	stats, err := fs.ClassDRMCardAMDGPUStats()
	if err != nil {
		return
	}
	var totalGpuUtil float64
	var gpuCount int

	// Export metrics for each card
	for _, card := range stats {

		// Edge case where we have no physical GPU
		if card.MemoryVRAMSize == 0 {
			continue
		}

		ch <- prometheus.MustNewConstMetric(
			gc.gpuBusyPercentDesc,
			prometheus.GaugeValue,
			float64(card.GPUBusyPercent),
			card.Name, card.UniqueID,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuGTTSizeDesc,
			prometheus.GaugeValue,
			float64(card.MemoryGTTSize),
			card.Name, card.UniqueID,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuGTTUsedDesc,
			prometheus.GaugeValue,
			float64(card.MemoryGTTUsed),
			card.Name, card.UniqueID,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuVisibleVRAMSizeDesc,
			prometheus.GaugeValue,
			float64(card.MemoryVisibleVRAMSize),
			card.Name, card.UniqueID,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuVisibleVRAMUsedDesc,
			prometheus.GaugeValue,
			float64(card.MemoryVisibleVRAMUsed),
			card.Name, card.UniqueID,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuMemoryVRAMSizeDesc,
			prometheus.GaugeValue,
			float64(card.MemoryVRAMSize),
			card.Name, card.UniqueID,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuMemoryVRAMUsedDesc,
			prometheus.GaugeValue,
			float64(card.MemoryVRAMUsed),
			card.Name, card.UniqueID,
		)

		gpuUtil := 0.7*float64(card.GPUBusyPercent) + 0.3*((float64(card.MemoryVisibleVRAMUsed)/float64(card.MemoryVRAMSize))*100)
		totalGpuUtil += gpuUtil
		gpuCount++
	}

	if gpuCount == 0 {
		return
	}

	avgGpuUtil := float64(totalGpuUtil) / float64(gpuCount)
	// Save for use in score.go
	SharedGpuUtil = (avgGpuUtil / 100)

	ch <- prometheus.MustNewConstMetric(
		gc.gpuAverageUtilizationDesc,
		prometheus.GaugeValue,
		avgGpuUtil,
	)

}
