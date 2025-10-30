package utility

import (
	"sync"

	"github.com/prometheus/procfs/sysfs"
)

var (
	linkSpeedsOnce sync.Once
	linkSpeeds     map[string]int64
)

func GetLinkSpeeds() map[string]int64 {
	linkSpeedsOnce.Do(func() {
		speeds := make(map[string]int64)
		fs, _ := sysfs.NewFS("/sys")
		netDevs, _ := fs.NetClass()

		for name, dev := range netDevs {
			if dev.Speed != nil && *dev.Speed > 0 {
				// Convert mb/s to B/s
				speeds[name] = *dev.Speed * 125000
			}
		}
		linkSpeeds = speeds
	})
	return linkSpeeds
}
