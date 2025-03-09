package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/datastructures"
)



func process_file_entry(basedir string, entry fs.FileInfo, c chan commons.File) {
	hash_channel := make(chan string)
	size_channel := make(chan commons.FileSize)
	go commons.Hash_file(basedir, entry.Name(), hash_channel)
	go commons.Get_human_reabable_size(entry.Size(), size_channel)

	file_size_info := <- size_channel 
	hash := <- hash_channel

	output := commons.File{
		Name: filepath.Join(basedir, entry.Name()),
		Size: file_size_info,
		Hash: hash,
	}

	c <- output
}

func display_file_info_from_channel(file_channel chan commons.File, file_count int) {
	for i := 0; i < file_count; i++ {
		data := <-file_channel
		fmt.Printf(
			"file: %-55s %6d %2s    %v\n",
			data.Name, data.Size.Value, data.Size.Unit, data.Hash,
		)
	}
}

func main() {
	var basedir string

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	directories_stack := datastructures.Stack{}
	datastructures.Push_into_stack(&directories_stack, basedir);

	for !datastructures.Is_stack_empty(&directories_stack) {
		current_dir, stack_err := datastructures.Get_stack_top_element(&directories_stack)

		if stack_err != nil {
			log.Fatal(stack_err)
		}

		datastructures.Pop_from_stack(&directories_stack)

		entries, read_dir_err := os.ReadDir(current_dir)
	
		if read_dir_err != nil {
			log.Fatal(read_dir_err)
		}

		file_channel := make(chan commons.File)
		file_counter := 0
	
		for _, entry := range entries {
			entry_info, file_info_err := entry.Info()

			if file_info_err != nil {
				log.Fatal(file_info_err)
			}

			if(entry_info.Mode().IsRegular()) {
				go process_file_entry(current_dir, entry_info, file_channel)
				file_counter++
			} else {
				datastructures.Push_into_stack(
					&directories_stack, 
					filepath.Join(current_dir, entry.Name()),
				)
			}
		}

		go display_file_info_from_channel(file_channel, file_counter)
	}
}