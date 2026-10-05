package sysinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

var (
	startTime      = time.Now()
	mu             sync.RWMutex
	lastCPU        float64
	lastPerCoreCPU []float64
)

type SystemInfo struct {
	CPUPercent     float64   `json:"cpu_percent"`
	CPUPerCore     []float64 `json:"cpu_per_core,omitempty"`
	RAMUsedBytes   uint64    `json:"ram_used_bytes"`
	RAMTotalBytes  uint64    `json:"ram_total_bytes"`
	RAMPercent     float64   `json:"ram_percent"`
	DiskUsedBytes  uint64    `json:"disk_used_bytes"`
	DiskTotalBytes uint64    `json:"disk_total_bytes"`
	DiskPercent    float64   `json:"disk_percent"`
	UptimeSeconds  uint64    `json:"uptime_seconds"`
	UptimeHuman    string    `json:"uptime_human"`
	OS             string    `json:"os"`
	Arch           string    `json:"arch"`
	Hostname       string    `json:"hostname"`
	NumCPU         int       `json:"num_cpu"`
	Goroutines     int       `json:"goroutines"`
	GoVersion      string    `json:"go_version"`
}

func init() {
	// Start continuous background CPU sampler matching htop's 1-second sampling window
	go startCPUSampler()
}

func startCPUSampler() {
	// Initial baseline reading
	_, _ = cpu.Percent(100*time.Millisecond, true)

	for {
		// Clean 1-second sampling across all logical cores (identical to htop)
		perCores, err := cpu.Percent(1*time.Second, true)
		if err == nil && len(perCores) > 0 {
			var total float64
			for _, p := range perCores {
				total += p
			}
			avg := total / float64(len(perCores))

			mu.Lock()
			lastCPU = avg
			lastPerCoreCPU = make([]float64, len(perCores))
			copy(lastPerCoreCPU, perCores)
			mu.Unlock()
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

// readMemInfoLinux calculates memory exactly matching htop from /proc/meminfo
// In htop: usedMem = MemTotal - MemFree - Buffers - Cached - SReclaimable + Shmem
func readMemInfoLinux() (uint64, uint64, float64, bool) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, 0, 0, false
	}

	var memTotal, memFree, buffers, cached, sReclaimable, shmem uint64
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		valFields := strings.Fields(parts[1])
		if len(valFields) == 0 {
			continue
		}
		valKB, _ := strconv.ParseUint(valFields[0], 10, 64)
		valBytes := valKB * 1024 // /proc/meminfo reports values in kB

		switch key {
		case "MemTotal":
			memTotal = valBytes
		case "MemFree":
			memFree = valBytes
		case "Buffers":
			buffers = valBytes
		case "Cached":
			cached = valBytes
		case "SReclaimable":
			sReclaimable = valBytes
		case "Shmem":
			shmem = valBytes
		}
	}

	if memTotal == 0 {
		return 0, 0, 0, false
	}

	// Exact htop memory formula (LinuxProcessList.c):
	// usedMem = MemTotal - MemFree - Buffers - (Cached + SReclaimable - Shmem)
	// which equals: MemTotal - MemFree - Buffers - Cached - SReclaimable + Shmem
	var nonUsed uint64 = memFree + buffers + cached + sReclaimable
	var used uint64
	if memTotal+shmem > nonUsed {
		used = (memTotal + shmem) - nonUsed
	} else {
		used = memTotal - memFree
	}

	pct := (float64(used) / float64(memTotal)) * 100.0
	return used, memTotal, pct, true
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

	// Memory info - use exact htop calculation on Linux
	if runtime.GOOS == "linux" {
		if used, total, pct, ok := readMemInfoLinux(); ok {
			info.RAMUsedBytes = used
			info.RAMTotalBytes = total
			info.RAMPercent = pct
		}
	}

	// Fallback to gopsutil if not on Linux or if /proc/meminfo wasn't readable
	if info.RAMTotalBytes == 0 {
		if vMem, err := mem.VirtualMemory(); err == nil {
			info.RAMUsedBytes = vMem.Used
			info.RAMTotalBytes = vMem.Total
			info.RAMPercent = vMem.UsedPercent
		} else {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			info.RAMUsedBytes = m.Alloc
			info.RAMTotalBytes = m.Sys
			if m.Sys > 0 {
				info.RAMPercent = float64(m.Alloc) / float64(m.Sys) * 100
			}
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

	// CPU percentage from background sampler
	mu.RLock()
	info.CPUPercent = lastCPU
	if len(lastPerCoreCPU) > 0 {
		info.CPUPerCore = make([]float64, len(lastPerCoreCPU))
		copy(info.CPUPerCore, lastPerCoreCPU)
	}
	mu.RUnlock()

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
