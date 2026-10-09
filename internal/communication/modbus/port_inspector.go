package modbus

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"datalogger/pkg/sysinfo"
)

// InspectSerialPort investigates the physical OS device path, symlink integrity,
// process permissions, ownership, and current system available ports.
func InspectSerialPort(configuredPort string) PortEvidence {
	evidence := PortEvidence{
		ConfiguredPort:       configuredPort,
		ResolvedPort:         configuredPort,
		SystemAvailablePorts: sysinfo.GetAvailableSerialPorts(),
	}

	if strings.TrimSpace(configuredPort) == "" {
		evidence.Exists = false
		evidence.CanAccess = false
		return evidence
	}

	// Bypass OS file checks for in-memory unit tests using mock serial pipes
	if strings.HasPrefix(configuredPort, "COM_TEST_") || strings.HasPrefix(configuredPort, "VIRTUAL_") {
		evidence.Exists = true
		evidence.CanAccess = true
		evidence.ResolvedPort = configuredPort
		return evidence
	}

	// 1. Check if configured port is a symlink or exists in filesystem
	lstat, lerr := os.Lstat(configuredPort)
	if lerr != nil {
		if os.IsNotExist(lerr) {
			evidence.Exists = false
			evidence.CanAccess = false

			// Check if it might be a COM port on Windows that exists in registry/device list
			if runtime.GOOS == "windows" {
				for _, p := range evidence.SystemAvailablePorts {
					if strings.EqualFold(p, configuredPort) {
						evidence.Exists = true
						evidence.ResolvedPort = p
						break
					}
				}
			}
			return evidence
		}
		if os.IsPermission(lerr) {
			evidence.Exists = true
			evidence.CanAccess = false
			evidence.PermissionError = lerr.Error()
			return evidence
		}
	}

	evidence.Exists = true

	// 2. Symlink Inspection (common on Linux Raspberry Pi /dev/serial/by-id or by-path)
	if lstat != nil && (lstat.Mode()&os.ModeSymlink != 0) {
		evidence.IsSymlink = true
		target, err := filepath.EvalSymlinks(configuredPort)
		if err == nil {
			evidence.SymlinkTarget = target
			if _, statErr := os.Stat(target); statErr == nil {
				evidence.SymlinkTargetExists = true
				evidence.ResolvedPort = target
			} else {
				evidence.SymlinkTargetExists = false
			}
		} else {
			// Symlink cannot be resolved (broken symlink)
			rawTarget, _ := os.Readlink(configuredPort)
			evidence.SymlinkTarget = rawTarget
			evidence.SymlinkTargetExists = false
		}
	}

	// If broken symlink: target does not exist
	if evidence.IsSymlink && !evidence.SymlinkTargetExists {
		evidence.CanAccess = false
		return evidence
	}

	// 3. Port Ownership Check (within Datalogger process)
	targetToCheck := evidence.ResolvedPort
	if IsPortOwnedByBus(targetToCheck) || IsPortOwnedByBus(configuredPort) {
		evidence.AdapterOwned = true
		evidence.CanAccess = true
		return evidence
	}

	// 4. File Permission & Busy Check (if not already opened by bus)
	if runtime.GOOS != "windows" {
		f, err := os.OpenFile(targetToCheck, os.O_RDWR, 0)
		if err != nil {
			if os.IsPermission(err) {
				evidence.CanAccess = false
				evidence.PermissionError = fmt.Sprintf("Permission denied on %s: %v", targetToCheck, err)
			} else if strings.Contains(strings.ToLower(err.Error()), "busy") ||
				strings.Contains(strings.ToLower(err.Error()), "temporarily unavailable") {
				evidence.IsBusy = true
				evidence.CanAccess = false
				evidence.BusyReason = fmt.Sprintf("Port %s is locked or busy: %v", targetToCheck, err)
			} else if os.IsNotExist(err) {
				evidence.Exists = false
				evidence.CanAccess = false
			} else {
				// Other OS error
				evidence.CanAccess = false
				evidence.PermissionError = err.Error()
			}
		} else {
			evidence.CanAccess = true
			_ = f.Close()
		}
	} else {
		// On Windows, if port exists in available ports, it's considered accessible unless COM open fails
		evidence.CanAccess = true
	}

	return evidence
}
