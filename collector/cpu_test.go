package collector

import (
	"testing"

	"github.com/prometheus/procfs"
)

func TestProcfs(t *testing.T) {
	fs, err := procfs.NewDefaultFS()

	if err != nil {
		return
	}

	stat, err := fs.Stat()
	if err != nil {
		return
	}

	cpuCount := len(stat.CPU)
	totalIdle := stat.CPUTotal.Idle

	println("cpuCount: %d, totalIdle: %d", cpuCount, totalIdle)
}
