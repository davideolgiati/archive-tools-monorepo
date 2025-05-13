package main

import (
	"archive-tools-monorepo/commons"
	_ "embed"
	"flag"
	"io/fs"
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

func check_if_file_is_allowed(full_path *string) bool {
	return evaluate_object_properties(full_path) == file
}

func submit_file_thread_pool(file *fs.FileInfo, current_dir *string, channel chan<- FsObj) {
	channel <- FsObj{obj: *file, base_dir: *current_dir}
}

func main() {
	var start_directory string = ""
	var skip_empty bool = false

	flag.StringVar(&start_directory, "dir", "", "Scan starting point  directory")
	flag.BoolVar(&skip_empty, "no_empty", false, "Skip empty files during scan")
	flag.Parse()

	commons.Print_not_registered(main_ui, "Running version: %s", version)
	commons.Print_not_registered(main_ui, "Build timestamp: %s\n", buildts)

	output_file_heap := build_new_file_heap()

	file_entry_channel := make(chan FsObj)

	for w := 1; w <= runtime.NumCPU()*4; w++ {
		go file_process_thread_pool(output_file_heap, file_entry_channel)
	}

	file_callback_fn := func(file *fs.FileInfo, current_dir *string) {
		submit_file_thread_pool(file, current_dir, file_entry_channel)
	}

	walker := New_dir_walker(skip_empty)

	walker.Set_entry_point(start_directory)
	walker.Set_directory_filter_function(check_if_dir_is_allowed)
	walker.Set_file_filter_function(check_if_file_is_allowed)
	walker.Set_file_callback_function(file_callback_fn)

	walker.Walk()

	close(file_entry_channel)

	output_file_heap.collect()

	cleaned_heap_1 := build_duplicate_entries_heap(&output_file_heap.heap, true)
	cleaned_heap := build_duplicate_entries_heap(cleaned_heap_1, false)

	commons.Close_UI(main_ui)

	display_duplicate_file_info(cleaned_heap)
}
