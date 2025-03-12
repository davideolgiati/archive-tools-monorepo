package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/datastructures"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"errors"
)

func process_file_entry(basedir string, entry fs.FileInfo, file_stack *datastructures.Stack[commons.File]) {
	file_size_info := commons.Get_human_reabable_size(entry.Size())
	hash_channel := make(chan string)
	
	go commons.Hash_file(basedir, entry.Name(), hash_channel)
	
	hash := <- hash_channel

	output := commons.File{
		Name: filepath.Join(basedir, entry.Name()),
		Size: file_size_info,
		Hash: hash,
	}

	datastructures.Push_into_stack(file_stack, output)
}

func display_file_info_from_channel(file_stack *datastructures.Stack[commons.File]) {
	for !datastructures.Is_stack_empty(file_stack) {
		data := datastructures.Pop_from_stack(file_stack)

		fmt.Printf(
			"file: %-55s %6d %2s    %v\n",
			data.Name, data.Size.Value, data.Size.Unit, data.Hash,
		)
	}
}

func get_next_directory_from_stack(directories_stack *datastructures.Stack[string]) (string, []os.DirEntry, error) {
	current_dir := datastructures.Pop_from_stack(directories_stack)

	entries, read_dir_err := os.ReadDir(current_dir)

	if read_dir_err != nil {
		return "", nil, read_dir_err
	}

	return current_dir, entries, nil
}

func main() {
	var basedir string

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	directories_stack := datastructures.Stack[string]{}
	file_stack := datastructures.Stack[commons.File]{}
	
	datastructures.Push_into_stack(&directories_stack, basedir)

	for !datastructures.Is_stack_empty(&directories_stack) {
		current_dir, entries, err := get_next_directory_from_stack(&directories_stack)

		if err != nil {
			panic(err)
		}

		for _, entry := range entries {
			entry_info, file_info_err := entry.Info()

			if file_info_err != nil {
				panic(file_info_err)
			}

			info, err := os.Stat(path.Join(current_dir, entry.Name()))

			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					continue
				}
				panic(err)
			}

			if info.Mode().Perm()&0444 == 0444 {
				if entry_info.Mode().IsRegular() {
					go process_file_entry(current_dir, entry_info, &file_stack)
				} else if entry_info.Mode().IsDir() {
					datastructures.Push_into_stack(
						&directories_stack,
						filepath.Join(current_dir, entry.Name()),
					)
				}
			}
		}

		go display_file_info_from_channel(&file_stack)
	}
}
