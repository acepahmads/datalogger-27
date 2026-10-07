//go:build windows

package sysinfo

import (
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
