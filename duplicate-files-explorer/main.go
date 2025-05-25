package main

import (
	"archive-tools-monorepo/commons"
	_ "embed"
	"flag"
	"strings"
	"sync"
)

//go:embed semver.txt
var version string

//go:embed buildts.txt
var buildts string

var main_ui = commons.New_UI()

func filter[T comparable](ss []T, value T) (ret []T) {
	for _, s := range ss {
		if s != value {
			ret = append(ret, s)
		}
	}
	return
}

func main() {
	var start_directory string = ""
	var ignored_dir_user string = ""
	var skip_empty bool = false
	var fsobj_pool commons.WriteOnlyThreadPool[FsObj] = commons.WriteOnlyThreadPool[FsObj]{}

	output_channel := make(chan commons.File)
	output_wg := sync.WaitGroup{}

	flag.StringVar(&start_directory, "dir", "", "Scan starting point  directory")
	flag.StringVar(&ignored_dir_user, "skip_dirs", "", "Skip user defined directories during scan (separated by comma)")
	flag.BoolVar(&skip_empty, "no_empty", false, "Skip empty files during scan")
	flag.Parse()

	user_dirs := filter(strings.Split(ignored_dir_user, ","), "")

	commons.Print_not_registered(main_ui, "Running version: %s", version)
	commons.Print_not_registered(main_ui, "Build timestamp: %s", buildts)

	output_file_heap := build_new_file_heap(commons.SizeDescending)

	if output_file_heap == nil {
		panic("error wile creating new file heap object")
	}

	fsobj_pool.Init(get_file_process_thread_fn(&output_file_heap.hash_registry, output_channel))

	walker := New_dir_walker(skip_empty)

	if walker == nil {
		panic("error wile creating new file walker object")
	}

	output_wg.Add(1)

	go func() {
		defer output_wg.Done()
		for data := range output_channel {
			output_file_heap.heap.Push(data)
		}
	}()

	walker.Set_entry_point(start_directory)
	walker.Set_directory_filter_function(get_directory_filter_fn(&user_dirs))
	walker.Set_file_filter_function(check_if_file_is_allowed)
	walker.Set_file_callback_function(get_file_callback_fn(&fsobj_pool))

	walker.Walk()

	fsobj_pool.Sync_and_close()

	close(output_channel)

	output_wg.Wait()
	

	// TODO: questi mi piacerebbe trasformarli in reduce, ma non è banale
	// come sembra, ci devo lavorare
	cleaned_heap := build_duplicate_entries_heap(output_file_heap)

	display_duplicate_file_info(cleaned_heap)

	commons.Close_UI(main_ui)
}
