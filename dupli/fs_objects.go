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

type FilesystemObject interface {
	CanBeRead() bool
	Type() int
}

type File struct {
	infos fs.FileInfo
	path  string
}

func (f *File) CanBeRead() (bool, error) {
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

func (f *File) Type() (int, error) {
	if f.path == "" {
		return invalid, fmt.Errorf("%w: fullpath is empty", os.ErrInvalid)
	}

	obj, err := os.Lstat(f.path)
	if err != nil {
		return invalid, fmt.Errorf("%w", err)
	}

	switch {
	case obj.IsDir():
		return directory, nil
	case !commons.HasReadPermission(&obj):
		return invalid, nil
	case commons.IsDevice(&obj):
		return device, nil
	case commons.IsSocket(&obj):
		return socket, nil
	case commons.IsPipe(&obj):
		return pipe, nil
	case commons.IsSymbolicLink(&obj):
		return symlink, nil
	case obj.Mode().Perm().IsRegular():
		return file, nil
	default:
		return invalid, nil
	}
}

func processFileEntry(
	file *File,
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
		hash, err = commons.CalculateHash(file.path)
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
) (func(File) error, error) {
	if flyweight == nil {
		return nil, fmt.Errorf("%w: flyweight is a nil pointer", os.ErrInvalid)
	}

	return func(file File) error {
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

func (f *File) IsAllowed() (bool, error) {
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

func getFileCallback(tp *commons.WriteOnlyThreadPool[File]) func(file File) {
	if tp == nil {
		panic("threadpool is nil")
	}

	return func(file File) {
		err := tp.Submit(file)
		if err != nil {
			panic(err)
		}
	}
}
