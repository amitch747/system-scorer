package utility

import (
	"sync"

	"github.com/prometheus/procfs/sysfs"
)

var (
	gpuConfigOnce sync.Once
	gpuNode       bool
	gpuCount      int
)

func GetGPUConfig() (bool, int) {
	gpuConfigOnce.Do(func() {
		fs, err := sysfs.NewFS("/sys")
		if err != nil {
			gpuNode = false
			gpuCount = 0
			return
		}

		gpuStats, err := fs.ClassDRMCardAMDGPUStats()
		if err != nil || len(gpuStats) == 0 {
			gpuNode = false
			gpuCount = 0
			return
		}
		gpuNode = true
		gpuCount = len(gpuStats)
	})
	return gpuNode, gpuCount
}
