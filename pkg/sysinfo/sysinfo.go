package sysinfo

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	startTime = time.Now()
	mu        sync.RWMutex
	lastCPU   float64
	lastCheck time.Time
)

type SystemInfo struct {
	CPUPercent     float64 `json:"cpu_percent"`
	RAMUsedBytes   uint64  `json:"ram_used_bytes"`
	RAMTotalBytes  uint64  `json:"ram_total_bytes"`
	RAMPercent     float64 `json:"ram_percent"`
	DiskUsedBytes  uint64  `json:"disk_used_bytes"`
	DiskTotalBytes uint64  `json:"disk_total_bytes"`
	DiskPercent    float64 `json:"disk_percent"`
	UptimeSeconds  uint64  `json:"uptime_seconds"`
	UptimeHuman    string  `json:"uptime_human"`
	OS             string  `json:"os"`
	Arch           string  `json:"arch"`
	Hostname       string  `json:"hostname"`
	NumCPU         int     `json:"num_cpu"`
	Goroutines     int     `json:"goroutines"`
	GoVersion      string  `json:"go_version"`
}

// GetInfo collects system metrics safely with minimal CPU impact
func GetInfo() SystemInfo {
	info := SystemInfo{
		OS:         runtime.GOOS,
		Arch:       runtime.GOARCH,
		NumCPU:     runtime.NumCPU(),
		Goroutines: runtime.NumGoroutine(),
		GoVersion:  runtime.Version(),
	}

	hostname, err := os.Hostname()
	if err == nil {
		info.Hostname = hostname
	}

	// Calculate uptime
	uptimeSec := uint64(time.Since(startTime).Seconds())
	if hostInfo, err := host.Info(); err == nil && hostInfo.Uptime > 0 {
		info.UptimeSeconds = hostInfo.Uptime
	} else {
		info.UptimeSeconds = uptimeSec
	}
	info.UptimeHuman = FormatDuration(time.Duration(info.UptimeSeconds) * time.Second)

	// Memory info
	if vMem, err := mem.VirtualMemory(); err == nil {
		info.RAMUsedBytes = vMem.Used
		info.RAMTotalBytes = vMem.Total
		info.RAMPercent = vMem.UsedPercent
	} else {
		// Fallback to runtime memory stats
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		info.RAMUsedBytes = m.Alloc
		info.RAMTotalBytes = m.Sys
		if m.Sys > 0 {
			info.RAMPercent = float64(m.Alloc) / float64(m.Sys) * 100
		}
	}

	// Disk info (current working drive/partition)
	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
		if pwd, err := os.Getwd(); err == nil && len(pwd) >= 3 {
			diskPath = pwd[:3] // e.g. "D:\"
		}
	}
	if dUsage, err := disk.Usage(diskPath); err == nil {
		info.DiskUsedBytes = dUsage.Used
		info.DiskTotalBytes = dUsage.Total
		info.DiskPercent = dUsage.UsedPercent
	}

	// CPU percentage with rate-limiting (to avoid blocking or high CPU spikes on Raspberry Pi)
	mu.Lock()
	now := time.Now()
	if now.Sub(lastCheck) >= 1*time.Second {
		percentages, err := cpu.Percent(0, false)
		if err == nil && len(percentages) > 0 {
			lastCPU = percentages[0]
		}
		lastCheck = now
	}
	info.CPUPercent = lastCPU
	mu.Unlock()

	return info
}

func FormatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%dd %02dh %02dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%02dh %02dm %02ds", hours, minutes, seconds)
	}
	return fmt.Sprintf("%02dm %02ds", minutes, seconds)
}
