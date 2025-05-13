package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	_ "embed"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//go:embed semver.txt
var version string

//go:embed buildts.txt
var buildts string

var main_ui = commons.New_UI()

type FsObj struct {
	obj      fs.FileInfo
	base_dir string
}

func file_process_thread_pool(file_heap *FileHeap, in <-chan FsObj) {
	for obj := range in {
		process_file_entry(obj.base_dir, &obj.obj, file_heap)
	}
}

func main() {
	var basedir string = ""
	var fullpath string = ""
	var formatted_size commons.FileSize = commons.FileSize{}

	var file_seen int = 0
	var directories_seen int = 0
	var size_processed int64 = 0

	commons.Print_not_registered(main_ui, "Running version: %s", version)
	commons.Print_not_registered(main_ui, "Build timestamp: %s\n", buildts)

	commons.Register_new_line("directory-line", "Directories seen: %6d", main_ui)
	commons.Register_new_line("file-line", "Files seen: %12d", main_ui)
	commons.Register_new_line("size-line", "Processed: %10d %2s", main_ui)

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	directories_stack := ds.Stack[string]{}
	output_file_heap := build_new_file_heap()

	input := make(chan FsObj)

	for w := 1; w <= runtime.NumCPU() * 4; w++ {
		go file_process_thread_pool(output_file_heap, input)
	}

	ds.Push_into_stack(&directories_stack, basedir)

	for !ds.Is_stack_empty(&directories_stack) {
		current_dir := ds.Pop_from_stack(&directories_stack)
		entries, read_dir_err := os.ReadDir(current_dir)

		if read_dir_err != nil {
			continue
		}

		for _, entry := range entries {
			fullpath = filepath.Join(current_dir, entry.Name())

			if strings.Contains(fullpath, "/dev") || strings.Contains(fullpath, "/sys") || strings.Contains(fullpath, "/proc") {
				continue
			}

			if entry.IsDir() {
				directories_seen += 1
				ds.Push_into_stack(&directories_stack, fullpath)
			} else {
				obj, err := entry.Info()

				if err == nil && evaluate_object_properties(&obj, &fullpath) == file {
					file_seen += 1
					size_processed += obj.Size()
					input <- FsObj{obj: obj, base_dir: current_dir}
				}
			}
		}

		formatted_size = commons.Format_file_size(size_processed)

		commons.Print_to_line(main_ui, "directory-line", directories_seen)
		commons.Print_to_line(main_ui, "file-line", file_seen)
		commons.Print_to_line(main_ui, "size-line", formatted_size.Value, formatted_size.Unit)
	}

	for ds.Get_counter_value(&output_file_heap.pending_insert) > 0 {
		apply_back_pressure(&output_file_heap.pending_insert)
	}

	cleaned_heap_1 := build_duplicate_entries_heap(&output_file_heap.heap, true)
	cleaned_heap := build_duplicate_entries_heap(cleaned_heap_1, false)

	commons.Close_UI(main_ui)

	display_duplicate_file_info(cleaned_heap)
}
