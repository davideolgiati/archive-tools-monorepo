package commons

import (
	"io/fs"
	"os"
)

type Stats struct {
	fs.FileInfo
}

func (info *Stats) HasReadPermission() bool {
	if info == nil {
		return false
	}

	filePermissionsBits := info.Mode().Perm()

	userReadOk := filePermissionsBits&fs.FileMode(0o400) != fs.FileMode(0o000)
	groupReadOk := filePermissionsBits&fs.FileMode(0o040) != fs.FileMode(0o000)
	othersReadOk := filePermissionsBits&fs.FileMode(0o004) != fs.FileMode(0o000)

	return userReadOk || groupReadOk || othersReadOk
}

func (info *Stats) IsSymbolicLink() bool {
	if info == nil {
		return false
	}

	return info.Mode()&os.ModeSymlink != 0
}

func (info *Stats) IsDevice() bool {
	if info == nil {
		return false
	}

	return info.Mode()&os.ModeDevice == os.ModeDevice
}

func (info *Stats) IsSocket() bool {
	if info == nil {
		return false
	}

	return info.Mode()&os.ModeSocket == os.ModeSocket
}

func (info *Stats) IsPipe() bool {
	if info == nil {
		return false
	}

	return info.Mode()&os.ModeNamedPipe == os.ModeNamedPipe
}
