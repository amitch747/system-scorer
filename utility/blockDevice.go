package utility

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type BlockDeviceInfo struct {
	Type       string
	Rotational bool
}

var (
	deviceInfoOnce sync.Once
	deviceInfoMap  map[string]BlockDeviceInfo
)

func DetectBlockDevices() map[string]BlockDeviceInfo {
	devices := make(map[string]BlockDeviceInfo)

	entries, err := os.ReadDir("/sys/block")
	if err != nil {
		return devices
	}

	for _, entry := range entries {
		deviceName := entry.Name()

		if strings.HasPrefix(deviceName, "loop") ||
			strings.HasPrefix(deviceName, "ram") ||
			strings.HasPrefix(deviceName, "zram") ||
			strings.HasPrefix(deviceName, "dm-") {
			continue
		}

		rotationPath := filepath.Join("/sys/block", deviceName, "queue", "rotational")
		subsysPath := filepath.Join("/sys/block", deviceName, "device", "subsystem")

		data, err := os.ReadFile(rotationPath)
		if err != nil {
			continue
		}
		isRotation := strings.TrimSpace(string(data)) == "1"

		deviceType := "SSD" // default
		if isRotation {
			deviceType = "HDD"
		}

		link, err := os.Readlink(subsysPath)
		if err == nil && strings.Contains(link, "nvme") {
			deviceType = "NVMe"
		}

		devices[deviceName] = BlockDeviceInfo{
			Type:       deviceType,
			Rotational: isRotation,
		}
	}
	return devices
}

func GetBlockDeviceInfoMap() map[string]BlockDeviceInfo {
	deviceInfoOnce.Do(func() {
		deviceInfoMap = DetectBlockDevices()
	})
	return deviceInfoMap
}

func MatchBaseDevice(diskName string, blockDeviceInfoMap map[string]BlockDeviceInfo) (string, bool) {
	if _, ok := blockDeviceInfoMap[diskName]; ok {
		return diskName, true
	}

	for base := range deviceInfoMap {
		if strings.HasPrefix(diskName, base) {
			return base, true
		}
	}

	return "", false
}
