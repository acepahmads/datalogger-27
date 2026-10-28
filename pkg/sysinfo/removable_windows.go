//go:build windows
// +build windows

package sysinfo

import (
	"golang.org/x/sys/windows"
)

func detectWindowsRemovableDrives() []RemovableDriveInfo {
	var results []RemovableDriveInfo

	drivesMask, err := windows.GetLogicalDrives()
	if err != nil {
		return results
	}

	for i := 0; i < 26; i++ {
		if (drivesMask & (1 << i)) != 0 {
			driveLetter := string(rune('A' + i))
			rootPath := driveLetter + ":\\"
			rootPtr, err := windows.UTF16PtrFromString(rootPath)
			if err != nil {
				continue
			}

			// DRIVE_REMOVABLE = 2
			driveType := windows.GetDriveType(rootPtr)
			if driveType == windows.DRIVE_REMOVABLE {
				var volNameBuf [256]uint16
				var fsNameBuf [256]uint16
				var serialNumber uint32
				var maxComponentLen uint32
				var flags uint32

				_ = windows.GetVolumeInformation(
					rootPtr,
					&volNameBuf[0],
					uint32(len(volNameBuf)),
					&serialNumber,
					&maxComponentLen,
					&flags,
					&fsNameBuf[0],
					uint32(len(fsNameBuf)),
				)

				volumeName := windows.UTF16ToString(volNameBuf[:])
				if volumeName == "" {
					volumeName = "USB Drive (" + driveLetter + ":)"
				} else {
					volumeName = volumeName + " (" + driveLetter + ":)"
				}
				fileSystem := windows.UTF16ToString(fsNameBuf[:])

				var freeBytesAvailable uint64
				var totalNumberOfBytes uint64
				var totalNumberOfFreeBytes uint64
				_ = windows.GetDiskFreeSpaceEx(rootPtr, &freeBytesAvailable, &totalNumberOfBytes, &totalNumberOfFreeBytes)

				results = append(results, RemovableDriveInfo{
					Device:     driveLetter + ":",
					MountPoint: rootPath,
					VolumeName: volumeName,
					FileSystem: fileSystem,
					TotalBytes: totalNumberOfBytes,
					FreeBytes:  totalNumberOfFreeBytes,
					TotalHuman: FormatBytes(totalNumberOfBytes),
					FreeHuman:  FormatBytes(totalNumberOfFreeBytes),
				})
			}
		}
	}

	return results
}
