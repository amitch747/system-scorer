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

// /sys/class/drm/card*/device/mem_busy_percent

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
			"syscore_gpu_average_utilization_percent",
			"System average of gpu utilization",
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
	avgGpuUtil := float64(totalGpuUtil) / float64(gpuCount)

	ch <- prometheus.MustNewConstMetric(
		gc.gpuAverageUtilizationDesc,
		prometheus.GaugeValue,
		avgGpuUtil,
	)

}
