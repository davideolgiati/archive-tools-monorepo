package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func process_file_entry(
	basedir *string, 
	entry *fs.FileInfo, 
	file_stack *ds.Stack[commons.File],
	file_to_process_counter *ds.AtomicCounter,
) {
	ds.Increment(file_to_process_counter)
	
	hash_channel := make(chan string)
	go commons.Hash_file(*basedir, (*entry).Name(), hash_channel)

	size_info := commons.Get_human_reabable_size((*entry).Size())
	fullpath := filepath.Join(*basedir, (*entry).Name())
	
	file_stats := commons.File{
		Name: fullpath,
		Size: size_info,
	}
	
	file_stats.Hash = <-hash_channel
	ds.Push_into_stack(file_stack, file_stats)
	ds.Decrement(file_to_process_counter)
}

func display_file_info_from_channel(
	file_stack *ds.Stack[commons.File], 
	directories_stack *ds.Stack[string],
	file_to_process_counter *ds.AtomicCounter,
) {
	for ds.Get_counter_value(file_to_process_counter) > 0 || !ds.Is_stack_empty(directories_stack) {
		for !ds.Is_stack_empty(file_stack) {
			data := ds.Get_top_stack_element(file_stack)

			hash := data.Hash
			size := data.Size.Value
			unit := data.Size.Unit
			name := data.Name

			fmt.Printf("file: %s %4d %2s %s\n", hash, size, unit, name)
			ds.Pop_from_stack(file_stack)
		}
	}
}

func main() {
	var basedir string

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	directories_stack := ds.Stack[string]{}
	output_file_stack := ds.Stack[commons.File]{}
	file_to_process_counter := ds.Create_new_atomic_counter()

	ds.Push_into_stack(&directories_stack, basedir)
	
	go display_file_info_from_channel(&output_file_stack, &directories_stack, file_to_process_counter)

	for !ds.Is_stack_empty(&directories_stack) {
		for ds.Get_counter_value(file_to_process_counter) > 500 {
			time.Sleep(10 * time.Millisecond) 
		}
		
		current_dir := ds.Get_top_stack_element(&directories_stack)
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
				go process_file_entry(
					&current_dir, &entry_info, 
					&output_file_stack, 
					file_to_process_counter,
				)
			}
			
			if file_type.IsDir() {
				ds.Push_into_stack(&directories_stack, fullpath)
			}
		}

		ds.Pop_from_stack(&directories_stack)
	}
}
