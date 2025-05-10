package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

type FileHeap struct {
	heap           ds.Heap[commons.File]
	pending_insert ds.AtomicCounter
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

func process_file_entry(basedir string, entry *fs.FileInfo, file_heap *FileHeap) {
	file_stats := commons.File{
		Name: filepath.Join(basedir, (*entry).Name()),
		Size: (*entry).Size(),
		Hash: "",
	}

	file_stats.FormattedSize = commons.Get_human_reabable_size(file_stats.Size)

	if can_file_be_read(&file_stats.Name) {
		ds.Increment(&file_heap.pending_insert)
		ds.Push_into_heap(&file_heap.heap, &file_stats)
		ds.Decrement(&file_heap.pending_insert)
	}
}

func print_file_details_to_stdout(data *commons.File) {
	fmt.Printf(
		"file: %s %4d %2s %s\n", 
		data.Hash, 
		data.FormattedSize.Value, 
		data.FormattedSize.Unit,
		data.Name,
	)
}

func evaluate_object_properties(obj *fs.FileInfo, fullpath *string) int {
	switch {
	case commons.Is_file_symbolic_link(fullpath):
		return symlink
	case commons.Is_file_a_device(obj):
		return device
	case commons.Is_file_a_socket(obj):
		return socket
	case commons.Is_file_a_pipe(obj):
		return pipe
	case !commons.Current_user_has_read_right_on_file(obj):
		return invalid
	case (*obj).IsDir():
		return directory
	case (*obj).Mode().Perm().IsRegular():
		return file
	default:
		return invalid
	}
}
