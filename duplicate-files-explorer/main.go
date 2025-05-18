package main

import (
	"archive-tools-monorepo/commons"
	_ "embed"
	"flag"
	"sync"
)

//go:embed semver.txt
var version string

//go:embed buildts.txt
var buildts string

var main_ui = commons.New_UI()

func main() {
	var start_directory string = ""
	var ignored_dir_user string = ""
	var skip_empty bool = false

	var wg sync.WaitGroup

	flag.StringVar(&start_directory, "dir", "", "Scan starting point  directory")
	flag.StringVar(&ignored_dir_user, "skip_dirs", "", "Skip user defined directories during scan (separated by comma)")
	flag.BoolVar(&skip_empty, "no_empty", false, "Skip empty files during scan")
	flag.Parse()

	commons.Print_not_registered(main_ui, "Running version: %s", version)
	commons.Print_not_registered(main_ui, "Build timestamp: %s\n", buildts)

	output_file_heap := build_new_file_heap()

	if output_file_heap == nil || output_file_heap.heap == nil || output_file_heap.pending_insert == nil {
		panic("error wile creating new file heap object")
	}

	file_entry_channel := make(chan FsObj)

	for w := 1; w <= 1; w++ {
		wg.Add(1)
		go file_process_thread_pool(output_file_heap, file_entry_channel, &wg)
	}

	walker := New_dir_walker(skip_empty)

	if walker == nil {
		panic("error wile creating new file walker object")
	}

	walker.Set_entry_point(start_directory)
	walker.Set_directory_filter_function(get_directory_filter_fn(ignored_dir_user))
	walker.Set_file_filter_function(check_if_file_is_allowed)
	walker.Set_file_callback_function(get_file_callback_fn(file_entry_channel))

	walker.Walk()

	close(file_entry_channel)

	wg.Wait()

	if output_file_heap.pending_insert.Value() > 0 {
		panic("file heap collect() not working properly, pending_indert > 0")
	}

	// TODO: questi mi piacerebbe trasformarli in reduce, ma non è banale
	// come sembra, ci devo lavorare
	cleaned_heap := build_duplicate_entries_heap(output_file_heap.heap, true)

	display_duplicate_file_info(cleaned_heap)

	commons.Close_UI(main_ui)
}
