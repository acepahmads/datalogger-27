//go:build !windows

package sysinfo

import (
	"path/filepath"
	"sort"
)

// GetAvailableSerialPorts lists all available hardware/USB serial ports on Linux/macOS
func GetAvailableSerialPorts() []string {
	ports := make([]string, 0)
	patterns := []string{
		"/dev/ttyUSB*",
		"/dev/ttyACM*",
		"/dev/ttyS*",
	}
	for _, p := range patterns {
		matches, err := filepath.Glob(p)
		if err == nil {
			ports = append(ports, matches...)
		}
	}
	sort.Strings(ports)
	return ports
}
