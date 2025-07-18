package main

import (
	"fmt"
	"io/fs"
	"os"
	"strings"
	"sync"

	"archive-tools-monorepo/commons"
	datastructures "archive-tools-monorepo/dataStructures"
)

const (
	directory = iota
	file      = iota
	symlink   = iota
	device    = iota
	socket    = iota
	pipe      = iota
	invalid   = iota
)

var ignoredDir = [...]string{"/dev", "/run", "/proc", "/sys"}

type FilesystemObject struct {
	infos fs.FileInfo
	path  string
}

func (f *FilesystemObject) CanBeRead() (bool, error) {
	if f.path == "" {
		return false, fmt.Errorf("%w: fullpath is empty", os.ErrInvalid)
	}

	filePointer, err := os.Open(f.path)
	if err != nil {
		return false, fmt.Errorf("%w", err)
	}

	defer func() {
		err = filePointer.Close()
		if err != nil {
			panic(err)
		}
	}()

	return true, nil
}

func (f *FilesystemObject) Type() (int, error) {
	if f.path == "" {
		return invalid, fmt.Errorf("%w: fullpath is empty", os.ErrInvalid)
	}

	obj, err := os.Lstat(f.path)
	if err != nil {
		return invalid, fmt.Errorf("%w", err)
	}

	casted := commons.Stats{
		FileInfo: obj,
	}

	switch {
	case obj.IsDir():
		return directory, nil
	case !casted.HasReadPermission():
		return invalid, nil
	case casted.IsDevice():
		return device, nil
	case casted.IsSocket():
		return socket, nil
	case casted.IsPipe():
		return pipe, nil
	case casted.IsSymbolicLink():
		return symlink, nil
	case obj.Mode().Perm().IsRegular():
		return file, nil
	default:
		return invalid, nil
	}
}

func processFileEntry(
	file *FilesystemObject,
	fileChannel chan<- commons.File,
	flyweight *datastructures.Flyweight[string],
	sizeFilter *sync.Map,
) error {
	var err error

	if file == nil {
		return fmt.Errorf("%w: file is a nil pointer", os.ErrInvalid)
	}

	if flyweight == nil {
		return fmt.Errorf("%w: flyweight is a nil pointer", os.ErrInvalid)
	}

	canBeRead, err := file.CanBeRead()
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	if !canBeRead {
		return fmt.Errorf("%w: file can't be read", os.ErrInvalid)
	}

	hash := ""
	size := file.infos.Size()
	_, loaded := sizeFilter.LoadOrStore(size, true)

	if size < 5000000 && loaded {
		hash, err = commons.GetSHA1HashFromPath(file.path)
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	hashPointer, err := flyweight.Instance(hash)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	fileStats := commons.File{
		Name: file.path,
		Size: size,
		Hash: hashPointer,
	}

	fileChannel <- fileStats
	return nil
}

func getFileProcessWorker(
	flyweight *datastructures.Flyweight[string],
	fileChannel chan<- commons.File,
	sizeFilter *sync.Map,
) (func(FilesystemObject) error, error) {
	if flyweight == nil {
		return nil, fmt.Errorf("%w: flyweight is a nil pointer", os.ErrInvalid)
	}

	return func(file FilesystemObject) error {
		return processFileEntry(&file, fileChannel, flyweight, sizeFilter)
	}, nil
}

func checkIfDirIsAllowed(fullPath *string, userBlacklist *[]string) bool {
	if fullPath == nil {
		panic("fillePath is a nil pointer")
	}

	if userBlacklist == nil {
		panic("userBlacklist is a nil pointer")
	}

	allowed := true
	for index := range ignoredDir {
		allowed = allowed && !strings.Contains(*fullPath, ignoredDir[index])
	}

	for index := range *userBlacklist {
		allowed = allowed && !strings.Contains(*fullPath, (*userBlacklist)[index])
	}

	return allowed
}

func (f *FilesystemObject) IsAllowed() (bool, error) {
	if f.path == "" {
		return false, fmt.Errorf("%w: f.path is empty", os.ErrInvalid)
	}

	fileType, err := f.Type()
	if err != nil {
		return false, err
	}

	return fileType == file, nil
}

func getDirectoryFilter(userBlacklist *[]string) func(string) bool {
	if userBlacklist == nil {
		panic("user_dir is a nil pointer")
	}

	return func(fullPath string) bool {
		return checkIfDirIsAllowed(&fullPath, userBlacklist)
	}
}

func getFileCallback(tp *commons.WriteOnlyThreadPool[FilesystemObject]) func(file FilesystemObject) {
	if tp == nil {
		panic("threadpool is nil")
	}

	return func(file FilesystemObject) {
		err := tp.Submit(file)
		if err != nil {
			panic(err)
		}
	}
}
