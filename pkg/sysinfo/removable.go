package sysinfo

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// RemovableDriveInfo represents metadata for a detected USB flash drive / removable storage
type RemovableDriveInfo struct {
	Device     string `json:"device"`      // e.g. "E:" or "/dev/sdb1"
	MountPoint string `json:"mount_point"` // e.g. "E:\" or "/media/usb0"
	VolumeName string `json:"volume_name"` // e.g. "KINGSTON" or "USB_DRIVE"
	FileSystem string `json:"file_system"` // e.g. "FAT32", "NTFS", "exFAT"
	TotalBytes uint64 `json:"total_bytes"`
	FreeBytes  uint64 `json:"free_bytes"`
	TotalHuman string `json:"total_human"`
	FreeHuman  string `json:"free_human"`
}

// RemovableStorageStatus represents system status for removable storage devices
type RemovableStorageStatus struct {
	Detected bool                 `json:"detected"`
	Count    int                  `json:"count"`
	Drives   []RemovableDriveInfo `json:"drives"`
}

// FormatBytes formats bytes into human-readable representation
func FormatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// GetRemovableDrives discovers connected USB flash drives / removable storage
func GetRemovableDrives() RemovableStorageStatus {
	var drives []RemovableDriveInfo

	if runtime.GOOS == "windows" {
		drives = detectWindowsRemovableDrives()
	} else if runtime.GOOS == "linux" {
		drives = detectLinuxRemovableDrives()
	} else {
		// Generic fallback (e.g. darwin/unix)
		drives = detectGenericRemovableDrives()
	}

	return RemovableStorageStatus{
		Detected: len(drives) > 0,
		Count:    len(drives),
		Drives:   drives,
	}
}

// detectLinuxRemovableDrives parses Linux /proc/mounts and /sys/block to identify USB flash drives
func detectLinuxRemovableDrives() []RemovableDriveInfo {
	var results []RemovableDriveInfo

	mountData, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return results
	}

	lines := strings.Split(string(mountData), "\n")
	seenMounts := make(map[string]bool)

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		device := fields[0]
		mountPoint := fields[1]
		fsType := fields[2]

		// Skip pseudo/virtual filesystems
		if strings.HasPrefix(device, "none") || strings.HasPrefix(device, "tmpfs") ||
			strings.HasPrefix(device, "devtmpfs") || strings.HasPrefix(device, "sysfs") ||
			strings.HasPrefix(device, "proc") || strings.HasPrefix(device, "cgroup") {
			continue
		}

		isRemovable := false

		// Criteria 1: Mounted in /media or /mnt
		if strings.HasPrefix(mountPoint, "/media/") || strings.HasPrefix(mountPoint, "/mnt/") {
			isRemovable = true
		}

		// Criteria 2: Backing block device has removable bit set in /sys/block
		if strings.HasPrefix(device, "/dev/sd") || strings.HasPrefix(device, "/dev/mmcblk") {
			baseName := filepath.Base(device)
			// Strip partition number (e.g. sdb1 -> sdb, mmcblk0p1 -> mmcblk0)
			blockDev := strings.TrimRight(baseName, "0123456789")
			blockDev = strings.TrimSuffix(blockDev, "p")
			removablePath := filepath.Join("/sys/block", blockDev, "removable")
			if remData, err := os.ReadFile(removablePath); err == nil {
				if strings.TrimSpace(string(remData)) == "1" {
					isRemovable = true
				}
			}
		}

		if isRemovable && !seenMounts[mountPoint] {
			seenMounts[mountPoint] = true
			volName := filepath.Base(mountPoint)
			if volName == "." || volName == "/" {
				volName = filepath.Base(device)
			}

			// Estimate space if accessible
			results = append(results, RemovableDriveInfo{
				Device:     device,
				MountPoint: mountPoint,
				VolumeName: volName,
				FileSystem: fsType,
				TotalHuman: "Mounted USB",
				FreeHuman:  "Available",
			})
		}
	}

	return results
}

// detectGenericRemovableDrives provides fallback scanning for common /Volumes or /media mounts
func detectGenericRemovableDrives() []RemovableDriveInfo {
	var results []RemovableDriveInfo
	commonDirs := []string{"/Volumes", "/media"}
	for _, base := range commonDirs {
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				mp := filepath.Join(base, e.Name())
				results = append(results, RemovableDriveInfo{
					Device:     mp,
					MountPoint: mp,
					VolumeName: e.Name(),
					FileSystem: "Removable",
					TotalHuman: "External Storage",
					FreeHuman:  "Available",
				})
			}
		}
	}
	return results
}
