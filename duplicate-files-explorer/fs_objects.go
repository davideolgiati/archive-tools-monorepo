package main

import (
	"archive-tools-monorepo/commons"
	"io"
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

	buffer := make([]byte, 1)
	_, file_read_error := file_pointer.Read(buffer)

	return file_read_error == nil || file_read_error == io.EOF
}

func evaluate_object_properties(fullpath *string) int {

	if *fullpath == "" {
		panic("evaluate_object_properties - fullpath is empty")
	}

	obj, err := os.Lstat(*fullpath)

	if err != nil {
		return invalid
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

func process_file_entry(full_path *string, entry *fs.FileInfo, file_heap *FileHeap) {
	if file_heap == nil {
		panic("file_heap is not fully referenced")
	}

	if can_file_be_read(full_path) {
		file_stats := commons.File{
			Name:          *full_path,
			Size:          (*entry).Size(),
			Hash:          file_heap.hash_registry.Cache_reference(""),
			FormattedSize: commons.Format_file_size((*entry).Size()),
		}

		file_heap.heap.Push(file_stats)
	}
}

func get_file_process_thread_fn(file_heap *FileHeap) func(FsObj) {
	return func(obj FsObj) {
		process_file_entry(&obj.base_dir, &obj.obj, file_heap)
	}
}

func check_if_dir_is_allowed(full_path *string, user_defined_dir *[]string) bool {
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
	return evaluate_object_properties(&full_path) == file
}

func get_directory_filter_fn(user_dirs *[]string) func(full_path string) bool {
	return func(full_path string) bool {
		return check_if_dir_is_allowed(&full_path, user_dirs)
	}
}

func get_file_callback_fn(tp *commons.WriteOnlyThreadPool[FsObj]) func(file fs.FileInfo, current_dir string) {
	return func(file fs.FileInfo, current_dir string) {
		tp.Submit(FsObj{obj: file, base_dir: current_dir})
	}
}
