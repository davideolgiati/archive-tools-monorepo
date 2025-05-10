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
	var err error

	file_pointer, err := os.Open(*fullpath)

	if err != nil {
		return false
	}

	defer file_pointer.Close()

	r := bufio.NewReader(file_pointer)

	buf := make([]byte, 100)

	_, err = r.Read(buf)

	if err != nil && err != io.EOF {
		return false
	}

	return true
}

func process_file_entry(basedir string, entry *fs.FileInfo, file_heap *FileHeap) {
	file_size := (*entry).Size()
	fullpath := filepath.Join(basedir, (*entry).Name())

	if file_size <= 0 {
		return
	}

	if !can_file_be_read(&fullpath) {
		return
	}

	ds.Increment(&file_heap.pending_insert)

	file_stats := commons.File{
		Name: fullpath,
		Size: (*entry).Size(),
	}

	file_stats.FormattedSize = commons.Get_human_reabable_size((*entry).Size())
	file_stats.Hash = ""

	ds.Push_into_heap(&file_heap.heap, &file_stats)
	ds.Decrement(&file_heap.pending_insert)
}

func print_file_details_to_stdout(data *commons.File) {
	hash := data.Hash
	size := data.FormattedSize.Value
	unit := data.FormattedSize.Unit
	name := data.Name

	fmt.Printf("file: %s %4d %2s %s\n", hash, size, unit, name)
}

func evaluate_object_properties(obj *fs.FileInfo, fullpath *string) int {
	if commons.Is_file_symbolic_link(fullpath) {
		return symlink
	}

	if commons.Is_file_a_device(obj) {
		return device
	}

	if commons.Is_file_a_socket(obj) {
		return socket
	}

	if commons.Is_file_a_pipe(obj) {
		return pipe
	}

	if !commons.Current_user_has_read_right_on_file(obj) {
		return invalid
	}

	if (*obj).IsDir() {
		return directory
	}

	if (*obj).Mode().Perm().IsRegular() {
		return file
	}

	return invalid
}
