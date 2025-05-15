package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"bufio"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

func can_file_be_read(fullpath *string) bool {
	file_pointer, file_open_error := os.Open(*fullpath)

	if file_open_error != nil {
		return false
	}

	defer file_pointer.Close()

	buffered_reader := bufio.NewReader(file_pointer)
	buffer := make([]byte, 100)
	_, file_read_error := buffered_reader.Read(buffer)

	return file_read_error == nil || file_read_error == io.EOF
}

func evaluate_object_properties(fullpath *string) int {
	obj, err := os.Stat(*fullpath)

	if err != nil {
		return invalid
	}

	switch {
	case obj.IsDir():
		return directory
	case commons.Is_symbolic_link(fullpath):
		return symlink
	case commons.Is_a_device(&obj):
		return device
	case commons.Is_a_socket(&obj):
		return socket
	case commons.Is_a_pipe(&obj):
		return pipe
	case !commons.Check_read_rights_on_file(&obj):
		return invalid
	case obj.Mode().Perm().IsRegular():
		return file
	default:
		return invalid
	}
}

func process_file_entry(basedir string, entry *fs.FileInfo, file_heap *FileHeap) {
	full_path := filepath.Join(basedir, (*entry).Name())

	if can_file_be_read(&full_path) {
		ds.Increment(&file_heap.pending_insert)

		file_stats := commons.File{
			Name:          full_path,
			Size:          (*entry).Size(),
			Hash:          "",
			FormattedSize: commons.Format_file_size((*entry).Size()),
		}

		file_heap.heap.Push(&file_stats)

		ds.Decrement(&file_heap.pending_insert)
	}
}

func file_process_thread_pool(file_heap *FileHeap, in <-chan FsObj) {
	for obj := range in {
		process_file_entry(obj.base_dir, &obj.obj, file_heap)
	}
}

var ignored_dir = [...]string{"/dev", "/run", "/proc", "/sys"}

func check_if_dir_is_allowed(full_path *string, user_defined_dir *[]string) bool {
	allowed := true
	for index := range ignored_dir {
		allowed = allowed && !strings.Contains(*full_path, ignored_dir[index])
	}

	if user_defined_dir != nil {
		for index := range *user_defined_dir {
			allowed = allowed && !strings.Contains(*full_path, (*user_defined_dir)[index])
		}
	}

	return allowed
}

func check_if_file_is_allowed(full_path *string) bool {
	return evaluate_object_properties(full_path) == file
}

func submit_file_thread_pool(file *fs.FileInfo, current_dir *string, channel chan<- FsObj) {
	channel <- FsObj{obj: *file, base_dir: *current_dir}
}

func get_directory_filter_fn(ignored_dir_user string) func(full_path *string) bool {
	return func(full_path *string) bool {
		if ignored_dir_user == "" {
			return check_if_dir_is_allowed(full_path, nil)
		} else {
			user_dirs := strings.Split(ignored_dir_user, ",")
			return check_if_dir_is_allowed(full_path, &user_dirs)
		}
	}
}

func get_file_callback_fn(input_channel *chan FsObj) func(file *fs.FileInfo, current_dir *string) {
	return func(file *fs.FileInfo, current_dir *string) {
		submit_file_thread_pool(file, current_dir, *input_channel)
	}
}
