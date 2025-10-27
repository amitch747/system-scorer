package collector

import (
	"github.com/prometheus/client_golang/prometheus"

	"fmt"

	"github.com/prometheus/procfs/sysfs"
)

type AMDGPUCollector struct {
	fs                 sysfs.FS
	gpuBusyPercent     *prometheus.Desc
	gpuGTTSize         *prometheus.Desc
	gpuGTTUsed         *prometheus.Desc
	gpuVisibleVRAMSize *prometheus.Desc
	gpuVisibleVRAMUsed *prometheus.Desc
	gpuMemoryVRAMSize  *prometheus.Desc
	gpuMemoryVRAMUsed  *prometheus.Desc

	// gpuClockDesc       *prometheus.Desc // Maybe skip
	// gpuMemoryBandwidth *prometheus.Desc // Not available on all cards
	// gpuProcessCount    *prometheus.Desc
}

/*
	---fs.ClassDRMCardAMDGPUStats()---
	Name                          string // The card name.
	GPUBusyPercent                uint64 // How busy the GPU is as a percentage.
	MemoryGTTSize                 uint64 // The size of the graphics translation table (GTT) block in bytes.
	MemoryGTTUsed                 uint64 // The used amount of the graphics translation table (GTT) block in bytes.
	MemoryVisibleVRAMSize         uint64 // The size of visible VRAM in bytes.
	MemoryVisibleVRAMUsed         uint64 // The used amount of visible VRAM in bytes.
	MemoryVRAMSize                uint64 // The size of VRAM in bytes.
	MemoryVRAMUsed                uint64 // The used amount of VRAM in bytes.
	MemoryVRAMVendor              string // The VRAM vendor name.
	PowerDPMForcePerformanceLevel string // The current power performance level.
	UniqueID                      string // The unique ID of the GPU that will persist from machine to machine.

*/

func NewAMDGPUCollector() (*AMDGPUCollector, error) {
	fs, err := sysfs.NewFS("/sys")
	if err != nil {
		return nil, fmt.Errorf("failed to open sysfs: %w", err)
	}

	return &AMDGPUCollector{
		fs: fs,
		gpuBusyPercent: prometheus.NewDesc(
			"syscraper_gpu_busy_percent",
			"Percentage GPU is busy.",
			[]string{"card", "model"},
			nil,
		),
		gpuGTTSize: prometheus.NewDesc(
			"syscraper_gpu_gtt_size",
			"Size of GTT block in bytes.",
			[]string{"card", "model"},
			nil,
		),
		gpuGTTUsed: prometheus.NewDesc(
			"syscraper_gpu_gtt_used",
			"Used bytes of GTT block.",
			[]string{"card", "model"},
			nil,
		),
		gpuVisibleVRAMSize: prometheus.NewDesc(
			"syscraper_gpu_visible_vram_size",
			"Size of visible VRAM in bytes.",
			[]string{"card", "model"},
			nil,
		),
		gpuVisibleVRAMUsed: prometheus.NewDesc(
			"syscraper_gpu_visible_vram_used",
			"Used bytes of visible VRAM.",
			[]string{"card", "model"},
			nil,
		),
		gpuMemoryVRAMSize: prometheus.NewDesc(
			"syscraper_gpu_vram_size",
			"Size of VRAM in bytes.",
			[]string{"card", "model"},
			nil,
		),
		gpuMemoryVRAMUsed: prometheus.NewDesc(
			"syscraper_gpu_vram_used",
			"Used bytes of VRAM.",
			[]string{"card", "model"},
			nil,
		),
	}, nil
}

func (gc *AMDGPUCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(gc, ch)
}

func (gc *AMDGPUCollector) Collect(ch chan<- prometheus.Metric) {
	stats, err := gc.fs.ClassDRMCardAMDGPUStats()
	if err != nil {
		return
	}
	// Export metrics for each card
	for _, card := range stats {
		ch <- prometheus.MustNewConstMetric(
			gc.gpuBusyPercent,
			prometheus.GaugeValue,
			float64(card.GPUBusyPercent),
			card.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuGTTSize,
			prometheus.GaugeValue,
			float64(card.MemoryGTTSize),
			card.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuGTTUsed,
			prometheus.GaugeValue,
			float64(card.MemoryGTTUsed),
			card.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuVisibleVRAMSize,
			prometheus.GaugeValue,
			float64(card.MemoryVisibleVRAMSize),
			card.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuVisibleVRAMUsed,
			prometheus.GaugeValue,
			float64(card.MemoryVisibleVRAMUsed),
			card.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuMemoryVRAMSize,
			prometheus.GaugeValue,
			float64(card.MemoryVRAMSize),
			card.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			gc.gpuMemoryVRAMUsed,
			prometheus.GaugeValue,
			float64(card.MemoryVRAMUsed),
			card.Name,
		)
	}

	// /sys/class/drm/card*/device/mem_busy_percent

}
