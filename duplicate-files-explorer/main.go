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
) {
	for !ds.Is_stack_empty(file_stack) {
		data := ds.Pop_from_stack(file_stack)

		hash := data.Hash
		size := data.Size.Value
		unit := data.Size.Unit
		name := data.Name

		fmt.Printf("file: %s %4d %2s %s\n", hash, size, unit, name)
	}
}

func compute_back_pressure(queue_size int64) time.Duration {
	if(queue_size < 100) {
		return 0 * time.Millisecond
	}

	if(queue_size < 500) {
		return 2 * time.Millisecond
	}

	if(queue_size < 1000) {
		return 5 * time.Millisecond
	}

	return 10 * time.Millisecond
}

func main() {
	var basedir string
	saveCursorPosition := "\033[s"
    	clearLine := "\033[u\033[K"

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	directories_stack := ds.Stack[string]{}
	output_file_stack := ds.Stack[commons.File]{}
	file_to_process_counter := ds.Create_new_atomic_counter()

	file_seen := 0
	directories_seen := 0
	size_processed := int64(0)

	ds.Push_into_stack(&directories_stack, basedir)
	
	fmt.Print(saveCursorPosition)
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
				file_seen += 1
				size_processed += entry_info.Size()
				go process_file_entry(
					&current_dir, &entry_info, 
					&output_file_stack, 
					file_to_process_counter,
				)
			}
			
			if file_type.IsDir() {
				directories_seen += 1
				ds.Push_into_stack(&directories_stack, fullpath)
			}
		}

		formatted_size := commons.Get_human_reabable_size(size_processed)
		back_pressure := compute_back_pressure(ds.Get_counter_value(file_to_process_counter))

		fmt.Print(clearLine)
		fmt.Printf(
			"Seen %6d files in %6d directories (%3d %2s)", 
			file_seen, directories_seen, formatted_size.Value, 
			formatted_size.Unit,
		)
		time.Sleep(back_pressure) // backpressure
	}

	for ds.Get_counter_value(file_to_process_counter) > 0 {
		time.Sleep(10 * time.Millisecond)
	}

	display_file_info_from_channel(&output_file_stack)
}
