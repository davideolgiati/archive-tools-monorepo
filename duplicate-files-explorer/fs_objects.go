package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"io/fs"
	"os"
	"strings"
	"sync"
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

var ignored_dir = [...]string{"/dev", "/run", "/proc", "/sys"}

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
		panic("can_file_be_read - fullpath is empty")
	}

	file_pointer, file_open_error := os.Open(f.path)
	if file_open_error != nil {
		return false
	}

	defer file_pointer.Close()

	return true
}

func (f *File)Type() int {

	if f.path == "" {
		panic("evaluate_object_properties - fullpath is empty")
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

func process_file_entry(file *File, file_chan chan<- commons.File, flyweight *ds.Flyweight[string], size_filter *sync.Map) {
	var err error

	if file == nil {
		panic("file is a nil pointer")
	}

	if flyweight == nil {
		panic("flyweight is a nil pointer")
	}

	if file.CanBeRead() {
		hash := ""
		size := file.infos.Size()
		_, loaded := size_filter.LoadOrStore(size, true)

		if size < 5000000 && loaded {
			hash, err = commons.CalculateHash(file.path)
			if err != nil {
				panic(err)
			}
		}

		formatted_size, err := commons.FormatFileSize(file.infos.Size())

		if err != nil {
			panic(err)
		}

		file_stats := commons.File{
			Name:          file.path,
			Size:          size,
			Hash:          flyweight.Instance(hash),
			FormattedSize: formatted_size,
		}

		file_chan <- file_stats
	}
}

func getFileProcessWorker(flyweight *ds.Flyweight[string], fileChannel chan<- commons.File, sizeFilter *sync.Map) func(File) {
	if flyweight == nil {
		panic("flyweight is a nil pointer")
	}

	return func(file File) {
		process_file_entry(&file, fileChannel, flyweight, sizeFilter)
	}
}

func check_if_dir_is_allowed(full_path *string, user_defined_dir *[]string) bool {
	if full_path == nil {
		panic("full_path is a nil pointer")
	}

	if user_defined_dir == nil {
		panic("user_defined_dir is a nil pointer")
	}

	allowed := true
	for index := range ignored_dir {
		allowed = allowed && !strings.Contains(*full_path, ignored_dir[index])
	}

	for index := range *user_defined_dir {
		allowed = allowed && !strings.Contains(*full_path, (*user_defined_dir)[index])
	}

	return allowed
}

func (f *File) IsAllowed() bool {
	if f.path == "" {
		panic("full_path is empty")
	}

	return f.Type() == file
}

func get_directory_filter_fn(user_dirs *[]string) func(full_path string) bool {
	if user_dirs == nil {
		panic("user_dir is a nil pointer")
	}

	return func(full_path string) bool {
		return check_if_dir_is_allowed(&full_path, user_dirs)
	}
}

func get_file_callback_fn(tp *commons.WriteOnlyThreadPool[File]) func(file File) {
	if tp == nil {
		panic("threadpool is nil")
	}

	return func(file File) {
		tp.Submit(file)
	}
}
