package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"io/fs"
	"os"
	"strings"
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

type FsObj struct {
	obj      fs.FileInfo
	base_dir string
}

var ignored_dir = [...]string{"/dev", "/run", "/proc", "/sys"}

func can_file_be_read(fullpath *string) bool {
	if *fullpath == "" {
	    panic("can_file_be_read - fullpath is empty")
	}
    
	file_pointer, file_open_error := os.Open(*fullpath)
	if file_open_error != nil {
	    return false
	}
	defer file_pointer.Close()
    
	return true
}

func evaluate_object_properties(fullpath *string) int {

	if *fullpath == "" {
		panic("evaluate_object_properties - fullpath is empty")
	}

	obj, err := os.Lstat(*fullpath)

	if err != nil {
		panic(err)
	}

	switch {
	case obj.IsDir():
		return directory
	case !commons.Check_read_rights_on_file(&obj):
		return invalid
	case commons.Is_a_device(&obj):
		return device
	case commons.Is_a_socket(&obj):
		return socket
	case commons.Is_a_pipe(&obj):
		return pipe
	case commons.Is_symbolic_link(&obj):
		return symlink
	case obj.Mode().Perm().IsRegular():
		return file
	default:
		return invalid
	}
}

func process_file_entry(full_path *string, entry *fs.FileInfo, file_chan chan<- commons.File, flyweight *ds.Flyweight[string]) {
	if full_path == nil {
		panic("full_path is a nil pointer")
	}

	if entry == nil {
		panic("user_defined_dir is a nil pointer")
	}

	if flyweight == nil {
		panic("flyweight is a nil pointer")
	}

	if can_file_be_read(full_path) {
		file_stats := commons.File{
			Name:          *full_path,
			Size:          (*entry).Size(),
			Hash:          flyweight.Cache_reference(""),
			FormattedSize: commons.Format_file_size((*entry).Size()),
		}

		file_chan <- file_stats
	}
}

func get_file_process_thread_fn(flyweight *ds.Flyweight[string], file_chan chan<- commons.File) func(FsObj) {
	if flyweight == nil {
		panic("flyweight is a nil pointer")
	}
	
	return func(obj FsObj) {
		process_file_entry(&obj.base_dir, &obj.obj, file_chan, flyweight)
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

func check_if_file_is_allowed(full_path string) bool {
	if full_path == "" {
		panic("full_path is empty")
	}

	return evaluate_object_properties(&full_path) == file
}

func get_directory_filter_fn(user_dirs *[]string) func(full_path string) bool {
	if user_dirs == nil {
		panic("user_dir is a nil pointer")
	}

	return func(full_path string) bool {
		return check_if_dir_is_allowed(&full_path, user_dirs)
	}
}

func get_file_callback_fn(tp *commons.WriteOnlyThreadPool[FsObj]) func(file fs.FileInfo, current_dir string) {
	if tp == nil {
		panic("threadpool is nil")
	}

	return func(file fs.FileInfo, current_dir string) {
		tp.Submit(FsObj{obj: file, base_dir: current_dir})
	}
}
