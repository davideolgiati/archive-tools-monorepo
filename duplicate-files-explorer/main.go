package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	_ "embed"
	"flag"
	"strings"
	"sync"
)

//go:embed semver.txt
var version string

//go:embed buildts.txt
var buildts string

var ui = commons.New_UI()

func filter[T comparable](input []T, filter_value T) []T {
	var output []T = make([]T, 0)

	for _, str := range input {
		if str == filter_value {
			continue
		}

		output = append(output, str)
	}

	return output
}

func main() {
	start_directory := ""
	ignored_dir_user := ""
	skip_empty := false
	profile := false
	fsobj_pool := commons.WriteOnlyThreadPool[FsObj]{}
	profiler := commons.Profiler{}
	
	shared_registry := ds.Flyweight[string]{}
	output_file_heap := new_file_heap(commons.HashDescending, &shared_registry)

	output_channel := make(chan commons.File, 10000)
	output_wg := sync.WaitGroup{}

	flag.StringVar(&start_directory, "dir", "", "Scan starting point  directory")
	flag.StringVar(&ignored_dir_user, "skip_dirs", "", "Skip user defined directories during scan (separated by comma)")
	flag.BoolVar(&skip_empty, "no_empty", false, "Skip empty files during scan")
	flag.BoolVar(&profile, "profile", false, "Profile program performances")

	flag.Parse()

	if profile {
		ui.Toggle_silence()
		profiler.Start()
	}

	user_dirs := filter(strings.Split(ignored_dir_user, ","), "")

	ui.Print_not_registered("Running version: %s", version)
	ui.Print_not_registered("Build timestamp: %s", buildts)


	if output_file_heap == nil {
		panic("error wile creating new file heap object")
	}

	fsobj_pool.Init(get_file_process_thread_fn(output_file_heap.hash_registry, output_channel, &output_file_heap.size_filter))

	walker := New_dir_walker(skip_empty)

	if walker == nil {
		panic("error wile creating new file walker object")
	}

	output_wg.Add(1)

	go func() {
		for data := range output_channel {
			output_file_heap.heap.Push(data)
		}
		output_wg.Done()
	}()

	walker.Set_entry_point(start_directory)
	walker.Set_directory_filter_function(get_directory_filter_fn(&user_dirs))
	walker.Set_file_filter_function(check_if_file_is_allowed)
	walker.Set_file_callback_function(get_file_callback_fn(&fsobj_pool))
	walker.Set_directory_exploration_callback_function(fsobj_pool.Sync)

	walker.Walk()

	fsobj_pool.Sync_and_close()

	close(output_channel)

	output_wg.Wait()

	cleaned_heap := output_file_heap.filter_heap(commons.Equal, &shared_registry)
	display_duplicate_file_info(cleaned_heap)

	ui.Close()

	if profile {
		profiler.Collect()
	}
}
