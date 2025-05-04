package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type FileHeap struct {
	heap ds.Heap[commons.File]
	pending_insert ds.AtomicCounter
}

func can_file_be_read(fullpath *string) bool {
	var err error
	
	file_pointer, err := os.Open(*fullpath)

	if err != nil {
		panic(err)
	}

	defer file_pointer.Close()

	r := bufio.NewReader(file_pointer)

	buf := make([]byte, 100)

	_, err = r.Read(buf)

	return err == nil
}

func process_file_entry(basedir *string, entry *fs.FileInfo, file_heap *FileHeap) {
	file_size := (*entry).Size() 
	fullpath := filepath.Join(*basedir, (*entry).Name())
	
	if file_size <= 0 {
		return
	}

	if !can_file_be_read(&fullpath) {
		return
	}

	ds.Increment(&file_heap.pending_insert)
	
	hash_channel := make(chan string)
	file_size_channel := make(chan commons.FileSize)
	
	go commons.Hash_file(fullpath, true, hash_channel)
	go commons.Get_human_reabable_size_async((*entry).Size(), file_size_channel)

	file_stats := commons.File{
		Name:          fullpath,
		Size:          (*entry).Size(),
	}

	file_stats.FormattedSize = <- file_size_channel
	file_stats.Hash = <-hash_channel

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
