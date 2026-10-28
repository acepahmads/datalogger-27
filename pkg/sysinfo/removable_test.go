package sysinfo

import (
	"testing"
)

func TestGetRemovableDrives(t *testing.T) {
	status := GetRemovableDrives()
	t.Logf("Removable drives detected: %v, count: %d", status.Detected, status.Count)
	for i, d := range status.Drives {
		t.Logf("Drive %d: Device=%s, Mount=%s, Vol=%s, Free=%s / Total=%s",
			i+1, d.Device, d.MountPoint, d.VolumeName, d.FreeHuman, d.TotalHuman)
	}
}
