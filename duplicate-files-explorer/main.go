package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

var wg sync.WaitGroup

func process_file_entry(basedir *string, entry *fs.FileInfo, file_stack *ds.Stack[commons.File]) {
	wg.Add(1)

	file_size_info := commons.Get_human_reabable_size((*entry).Size())
	hash_channel := make(chan string)

	go commons.Hash_file(*basedir, (*entry).Name(), hash_channel)

	hash := <-hash_channel

	output := commons.File{
		Name: filepath.Join(*basedir, (*entry).Name()),
		Size: file_size_info,
		Hash: hash,
	}

	ds.Push_into_stack(file_stack, output)
	
	wg.Done()
}

func display_file_info_from_channel(file_stack *ds.Stack[commons.File]) {
	for !ds.Is_stack_empty(file_stack) {
		data := ds.Pop_from_stack(file_stack)
		hash := data.Hash
		size := data.Size.Value
		unit := data.Size.Unit
		name := data.Name

		fmt.Printf("file: %s %4d %2s %s\n", hash, size, unit, name)
	}
}

func main() {
	var basedir string

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	directories_stack := ds.Stack[string]{}
	file_stack := ds.Stack[commons.File]{}

	ds.Push_into_stack(&directories_stack, basedir)

	for !ds.Is_stack_empty(&directories_stack) {
		current_dir := ds.Pop_from_stack(&directories_stack)
		entries, read_dir_err := os.ReadDir(current_dir)

		if read_dir_err != nil {
			panic(read_dir_err)
		}

		for _, entry := range entries {
			entry_info, file_info_err := entry.Info()

			if file_info_err != nil {
				panic(file_info_err)
			}

			entry_name := entry_info.Name()
			file_type := entry_info.Mode()
			fullpath := filepath.Join(current_dir, entry_name)

			if !commons.Current_user_has_read_right_on_file(&fullpath) {
				continue
			}

			if file_type.IsRegular() {
				go process_file_entry(&current_dir, &entry_info, &file_stack)
			}
			
			if file_type.IsDir() {
				ds.Push_into_stack(&directories_stack, fullpath)
			}
		}
		
		go display_file_info_from_channel(&file_stack)
	}

	wg.Wait()
	display_file_info_from_channel(&file_stack)
}
