package main

import (
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

func (f *File) CanBeRead() bool {
	if f.path == "" {
		panic("CanBeRead - fullpath is empty")
	}

	filePointer, err := os.Open(f.path)
	if err != nil {
		return false
	}

	defer func() {
		err := filePointer.Close()
		if err != nil {
			panic(err)
		}
	}()

	return true
}

func (f *File) Type() int {
	if f.path == "" {
		panic("Type - fullpath is empty")
	}

	obj, err := os.Lstat(f.path)
	if err != nil {
		panic(err)
	}

	switch {
	case obj.IsDir():
		return directory
	case !commons.HasReadPermission(&obj):
		return invalid
	case commons.IsDevice(&obj):
		return device
	case commons.IsSocket(&obj):
		return socket
	case commons.IsPipe(&obj):
		return pipe
	case commons.IsSymbolicLink(&obj):
		return symlink
	case obj.Mode().Perm().IsRegular():
		return file
	default:
		return invalid
	}
}

func processFileEntry(file *File, fileChannel chan<- commons.File, flyweight *datastructures.Flyweight[string], sizeFilter *sync.Map) {
	var err error

	if file == nil {
		panic("file is a nil pointer")
	}

	if flyweight == nil {
		panic("flyweight is a nil pointer")
	}

	if !file.CanBeRead() {
		return
	}

	hash := ""
	size := file.infos.Size()
	_, loaded := sizeFilter.LoadOrStore(size, true)

	if size < 5000000 && loaded {
		hash, err = commons.CalculateHash(file.path)
		if err != nil {
			panic(err)
		}
	}

	formattedSize, err := commons.FormatFileSize(file.infos.Size())
	if err != nil {
		panic(err)
	}

	hashPointer, err := flyweight.Instance(hash)
	if err != nil {
		panic(err)
	}

	fileStats := commons.File{
		Name:          file.path,
		Size:          size,
		Hash:          hashPointer,
		FormattedSize: formattedSize,
	}

	fileChannel <- fileStats
}

func getFileProcessWorker(flyweight *datastructures.Flyweight[string], fileChannel chan<- commons.File, sizeFilter *sync.Map) func(File) {
	if flyweight == nil {
		panic("flyweight is a nil pointer")
	}

	return func(file File) {
		processFileEntry(&file, fileChannel, flyweight, sizeFilter)
	}
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

func (f *File) IsAllowed() bool {
	if f.path == "" {
		panic("f.path is empty")
	}

	return f.Type() == file
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
