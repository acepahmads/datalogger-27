//go:build windows

package sysinfo

import (
	"fmt"
	"sort"

	"golang.org/x/sys/windows/registry"
)

// GetAvailableSerialPorts lists all available hardware/virtual COM ports detected on Windows
func GetAvailableSerialPorts() []string {
	ports := make([]string, 0)
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DEVICEMAP\SERIALCOMM`, registry.QUERY_VALUE)
	if err != nil {
		return ports
	}
	defer k.Close()

	names, err := k.ReadValueNames(-1)
	if err != nil {
		return ports
	}

	for _, name := range names {
		val, _, err := k.GetStringValue(name)
		if err == nil && val != "" {
			ports = append(ports, val)
		}
	}
	sort.Strings(ports)
	return ports
}

// GetDetailedSerialPorts returns serial ports with metadata on Windows
func GetDetailedSerialPorts() []SerialPortDetail {
	ports := GetAvailableSerialPorts()
	results := make([]SerialPortDetail, 0, len(ports))
	for _, p := range ports {
		results = append(results, SerialPortDetail{
			Path:        p,
			RealDev:     p,
			Type:        "STANDARD",
			Description: fmt.Sprintf("Windows Serial Port (%s)", p),
			Recommended: true,
		})
	}
	return results
}
