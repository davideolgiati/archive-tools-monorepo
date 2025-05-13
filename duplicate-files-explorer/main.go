package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	_ "embed"
	"flag"
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
var ignored_dir = [...]string{"/dev", "/run", "/proc", "/sys"}

func check_if_dir_is_allowed(full_path *string) bool {
	allowed := true
	for index := range ignored_dir {
		allowed = allowed && !strings.Contains(*full_path, ignored_dir[index])
	}

	return allowed
}

func main() {
	var basedir string = ""
	var fullpath string = ""
	var formatted_size commons.FileSize = commons.FileSize{}

	var file_seen int = 0
	var directories_seen int = 0
	var size_processed int64 = 0
	var skip_empty bool = false

	commons.Print_not_registered(main_ui, "Running version: %s", version)
	commons.Print_not_registered(main_ui, "Build timestamp: %s\n", buildts)

	main_ui.Register_line("directory-line", "Directories seen: %6d")
	main_ui.Register_line("file-line", "Files seen: %12d")
	main_ui.Register_line("size-line", "Processed: %10d %2s")

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.BoolVar(&skip_empty, "no_empty", false, "Skip empty files during scan")
	flag.Parse()

	directories_stack := ds.Stack[string]{}
	output_file_heap := build_new_file_heap()

	input := make(chan FsObj)

	for w := 1; w <= runtime.NumCPU()*4; w++ {
		go file_process_thread_pool(output_file_heap, input)
	}

	directories_stack.Push(basedir)

	for !directories_stack.Empty() {
		current_dir := directories_stack.Pop()
		entries, read_dir_err := os.ReadDir(current_dir)

		if read_dir_err != nil {
			continue
		}

		for _, entry := range entries {
			fullpath = filepath.Join(current_dir, entry.Name())

			if !check_if_dir_is_allowed(&fullpath) {
				continue
			}

			if entry.IsDir() {
				directories_seen += 1
				directories_stack.Push(fullpath)
			} else {
				obj, err := entry.Info()

				if skip_empty && obj.Size() == 0 {
					continue
				}

				if err == nil && evaluate_object_properties(&obj, &fullpath) == file {
					file_seen += 1
					size_processed += obj.Size()
					input <- FsObj{obj: obj, base_dir: current_dir}
				}
			}
		}

		formatted_size = commons.Format_file_size(size_processed)

		main_ui.Update_line("directory-line", directories_seen)
		main_ui.Update_line("file-line", file_seen)
		main_ui.Update_line("size-line", formatted_size.Value, formatted_size.Unit)
	}

	close(input)

	output_file_heap.collect()

	cleaned_heap_1 := build_duplicate_entries_heap(&output_file_heap.heap, true)
	cleaned_heap := build_duplicate_entries_heap(cleaned_heap_1, false)

	commons.Close_UI(main_ui)

	display_duplicate_file_info(cleaned_heap)
}
