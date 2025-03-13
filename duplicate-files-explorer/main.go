package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/datastructures"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

func process_file_entry(
	basedir *string, 
	entry *fs.FileInfo, 
	file_stack *datastructures.Stack[commons.File],
	wg *sync.WaitGroup,
) {
	if *basedir == "" || basedir == nil {
		panic("basedir is invalid")
	}

	if entry == nil {
		panic("entry is invalid")
	}

	if file_stack == nil {
		panic("file_stack is invalid")
	}

	file_size_info := commons.Get_human_reabable_size((*entry).Size())
	hash_channel := make(chan string)

	go commons.Hash_file(*basedir, (*entry).Name(), hash_channel)

	hash := <-hash_channel

	if hash == "" {
		panic("basedir is empty")
	}

	output := commons.File{
		Name: filepath.Join(*basedir, (*entry).Name()),
		Size: file_size_info,
		Hash: hash,
	}

	if output.Name == "" {
		panic("output fullpath is empty")
	}

	if output.Hash == "" {
		panic("output hash is empty")
	}

	if output.Size.Unit == "" {
		panic("output size unit is empty")
	}

	datastructures.Push_into_stack(file_stack, output)
	
	wg.Done()
}

func display_file_info_from_channel(file_stack *datastructures.Stack[commons.File]) {
	for !datastructures.Is_stack_empty(file_stack) {
		data := datastructures.Pop_from_stack(file_stack)
		hash := data.Hash
		size := data.Size.Value
		unit := data.Size.Unit
		name := data.Name

		fmt.Printf("file: %s %4d %2s %s\n", hash, size, unit, name)
	}
}

func get_file_list_for_directory(current_dir *string) []os.DirEntry {
	entries, read_dir_err := os.ReadDir(*current_dir)

	if read_dir_err != nil {
		panic(read_dir_err)
	}

	return entries
}

func main() {
	var basedir string

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	directories_stack := datastructures.Stack[string]{}
	file_stack := datastructures.Stack[commons.File]{}
	var wg sync.WaitGroup

	datastructures.Push_into_stack(&directories_stack, basedir)

	for !datastructures.Is_stack_empty(&directories_stack) {
		current_dir := datastructures.Pop_from_stack(&directories_stack)
		entries := get_file_list_for_directory(&current_dir)

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
				wg.Add(1)
				go process_file_entry(&current_dir, &entry_info, &file_stack, &wg)
				go display_file_info_from_channel(&file_stack)
			}
			
			if file_type.IsDir() {
				datastructures.Push_into_stack(&directories_stack, fullpath)
			}
		}
	}

	wg.Wait()
	go display_file_info_from_channel(&file_stack)
}
