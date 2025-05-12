package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"bufio"
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

func build_new_file_heap() *FileHeap {
	file_heap := FileHeap{}

	ds.Set_compare_fn(&file_heap.heap, commons.Lower)
	file_heap.pending_insert = *ds.Build_new_atomic_counter()

	return &file_heap
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
	full_path := filepath.Join(basedir, (*entry).Name())

	if can_file_be_read(&full_path) {
		ds.Increment(&file_heap.pending_insert)

		file_stats := commons.File{
			Name: full_path,
			Size: (*entry).Size(),
			Hash: "",
			FormattedSize: commons.Format_file_size((*entry).Size()),
		}

		ds.Push_into_heap(&file_heap.heap, &file_stats)

		ds.Decrement(&file_heap.pending_insert)
	}
}

func evaluate_object_properties(obj *fs.FileInfo, fullpath *string) int {
	switch {
	case commons.Is_symbolic_link(fullpath):
		return symlink
	case commons.Is_a_device(obj):
		return device
	case commons.Is_a_socket(obj):
		return socket
	case commons.Is_a_pipe(obj):
		return pipe
	case !commons.Check_read_rights_on_file(obj):
		return invalid
	case (*obj).IsDir():
		return directory
	case (*obj).Mode().Perm().IsRegular():
		return file
	default:
		return invalid
	}
}
