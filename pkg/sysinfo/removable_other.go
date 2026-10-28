//go:build !windows
// +build !windows

package sysinfo

func detectWindowsRemovableDrives() []RemovableDriveInfo {
	return nil
}
