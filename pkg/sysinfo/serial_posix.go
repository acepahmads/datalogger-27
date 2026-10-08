//go:build !windows

package sysinfo

import (
	"fmt"
	"path/filepath"
	"sort"
)

// GetAvailableSerialPorts lists all available hardware/USB serial ports on Linux/macOS,
// placing persistent by-id and by-path ports first.
func GetAvailableSerialPorts() []string {
	details := GetDetailedSerialPorts()
	ports := make([]string, 0, len(details))
	for _, d := range details {
		ports = append(ports, d.Path)
	}
	return ports
}

// GetDetailedSerialPorts lists all available serial ports on Linux with by-id, by-path, and standard /dev/tty*
func GetDetailedSerialPorts() []SerialPortDetail {
	var results []SerialPortDetail
	seenReal := make(map[string]bool)

	// 1. Scan persistent by-id (based on USB chip unique serial number)
	if byIDMatches, err := filepath.Glob("/dev/serial/by-id/*"); err == nil && len(byIDMatches) > 0 {
		sort.Strings(byIDMatches)
		for _, p := range byIDMatches {
			realTarget, err := filepath.EvalSymlinks(p)
			if err != nil {
				realTarget = p
			}
			baseName := filepath.Base(p)
			results = append(results, SerialPortDetail{
				Path:        p,
				RealDev:     realTarget,
				Type:        "BY_ID",
				Description: fmt.Sprintf("Persistent by-ID (%s ➔ %s)", baseName, filepath.Base(realTarget)),
				Recommended: true,
			})
			seenReal[realTarget] = true
		}
	}

	// 2. Scan persistent by-path (based on physical USB socket port on Raspberry Pi)
	if byPathMatches, err := filepath.Glob("/dev/serial/by-path/*"); err == nil && len(byPathMatches) > 0 {
		sort.Strings(byPathMatches)
		for _, p := range byPathMatches {
			realTarget, err := filepath.EvalSymlinks(p)
			if err != nil {
				realTarget = p
			}
			results = append(results, SerialPortDetail{
				Path:        p,
				RealDev:     realTarget,
				Type:        "BY_PATH",
				Description: fmt.Sprintf("Persistent by-Path (Physical USB Socket ➔ %s)", filepath.Base(realTarget)),
				Recommended: true,
			})
			seenReal[realTarget] = true
		}
	}

	// 3. Scan standard direct ports (/dev/ttyUSB*, /dev/ttyACM*, /dev/ttyS*)
	patterns := []string{
		"/dev/ttyUSB*",
		"/dev/ttyACM*",
		"/dev/ttyS*",
	}
	for _, pat := range patterns {
		matches, err := filepath.Glob(pat)
		if err == nil {
			sort.Strings(matches)
			for _, m := range matches {
				isAlreadyCovered := seenReal[m]
				desc := "Direct Hardware Port"
				if isAlreadyCovered {
					desc = fmt.Sprintf("Direct Port %s (Persistent by-id/by-path available)", filepath.Base(m))
				}
				results = append(results, SerialPortDetail{
					Path:        m,
					RealDev:     m,
					Type:        "STANDARD",
					Description: desc,
					Recommended: !isAlreadyCovered,
				})
			}
		}
	}

	return results
}

