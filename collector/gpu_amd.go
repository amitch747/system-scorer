package collector

import "github.com/prometheus/client_golang/prometheus"

type AMDGPUCollector struct {
	gpuBusyDesc        *prometheus.Desc
	gpuVRAMDesc        *prometheus.Desc
	gpuClockDesc       *prometheus.Desc // Maybe skip
	gpuMemoryBandwidth *prometheus.Desc // Not available on all cards
	gpuProcessCount    *prometheus.Desc
}

func NewAMDGPUCollector() *AMDGPUCollector {
	return &AMDGPUCollector{
		gpuBusyDesc: prometheus.NewDesc(
			"amdgpu_gpu_busy_percent",
			"Current GPU utilization.",
			[]string{"index", "model"},
			nil,
		),
		gpuVRAMDesc: prometheus.NewDesc(
			"amdgpu_vram_used_percent",
			"Current VRAM utilization.",
			[]string{"index", "model"},
			nil,
		),
		gpuClockDesc: prometheus.NewDesc(
			"amdgpu_clock_frequency",
			"Curent GPU clock freq.",
			[]string{"index", "model"},
			nil,
		),
		gpuMemoryBandwidth: prometheus.NewDesc(
			"amdgpu_memory_bandwidth",
			"Current GPU memory bandwidth usage.",
			[]string{"index", "model"},
			nil,
		),
		gpuProcessCount: prometheus.NewDesc(
			"amdgpu_process_count",
			"Number of running processes on this GPU.",
			[]string{"index", "model"},
			nil,
		),
	}
}
