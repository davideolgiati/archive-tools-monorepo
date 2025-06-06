package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"errors"
	"io/fs"
	"os"
	"path"
)

type dirWalker struct {
	directory_filter_function func(string) bool
	file_filter_function      func(string) bool
	file_callback_function    func(fs.FileInfo, string)
	current_directory         string
	current_file              string
	directories               ds.PriorityQueue[string]
	size_processed            int64
	file_seen                 int
	directories_seen          int
	current_depth             uint
	skip_empty                bool
}

func New_dir_walker(skip_empty bool) *dirWalker {
	walker := dirWalker{}

	walker.directories = ds.PriorityQueue[string]{}
	walker.directories.Init()

	walker.file_seen = 0
	walker.directories_seen = 0
	walker.size_processed = 0
	walker.directory_filter_function = nil
	walker.file_filter_function = nil
	walker.file_callback_function = nil
	walker.skip_empty = skip_empty
	walker.current_directory = ""
	walker.current_file = ""

	return &walker
}

func (walker *dirWalker) Set_entry_point(directory string) {
	walker.directories.Push(directory, 0)
}

func (walker *dirWalker) Set_directory_filter_function(filter_fn func(string) bool) {
	walker.directory_filter_function = filter_fn
}

func (walker *dirWalker) Set_file_filter_function(filter_fn func(string) bool) {
	walker.file_filter_function = filter_fn
}

func (walker *dirWalker) Set_file_callback_function(callback func(fs.FileInfo, string)) {
	walker.file_callback_function = callback
}

func (walker *dirWalker) Walk() {
	var formatted_size commons.FileSize

	ui.Register_line("directory-line", "Directories seen: %6d")
	ui.Register_line("file-line", "Files seen: %12d")
	ui.Register_line("size-line", "Processed: %10d %2s")

	for !walker.directories.Empty() {
		walker.current_depth, walker.current_directory = walker.directories.Pop()
		objects, read_dir_err := os.ReadDir(walker.current_directory)

		if read_dir_err == nil {
			walker.process_directory_objects(&objects)

			formatted_size = commons.Format_file_size(walker.size_processed)

			ui.Update_line("directory-line", walker.directories_seen)
			ui.Update_line("file-line", walker.file_seen)
			ui.Update_line("size-line", formatted_size.Value, *formatted_size.Unit)
		} else if !errors.Is(read_dir_err, os.ErrPermission) {
			panic(read_dir_err)
		}
	}
}

func (walker *dirWalker) process_directory_objects(objects *[]os.DirEntry) {
	for _, obj := range *objects {
		walker.current_file = path.Join(walker.current_directory, obj.Name())

		if obj.IsDir() {
			walker.process_directory(&walker.current_file)
		} else {
			walker.process_file(&obj)
		}
	}
}

func (walker *dirWalker) process_directory(directory *string) {
	if !walker.directory_filter_function(*directory) {
		return
	}

	walker.directories_seen += 1
	walker.directories.Push(*directory, walker.current_depth)
}

func (walker *dirWalker) process_file(obj *os.DirEntry) {
	if !walker.file_filter_function(walker.current_file) {
		return
	}

	file_entry, err := (*obj).Info()

	if err != nil {
		panic(err)
	}

	if walker.skip_empty && file_entry.Size() == 0 {
		return
	}

	walker.file_seen += 1
	walker.size_processed += file_entry.Size()
	walker.file_callback_function(file_entry, walker.current_file)
}
