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

var ui = commons.NewUI()

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
	profiler := commons.Profiler{}

	var fsobjPool *commons.WriteOnlyThreadPool[File]

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
		ui.ToggleSilence()
		profiler.Start()
	}

	user_dirs := filter(strings.Split(ignored_dir_user, ","), "")

	ui.Println("Running version: %s", version)
	ui.Println("Build timestamp: %s", buildts)

	if output_file_heap == nil {
		panic("error wile creating new file heap object")
	}

	workerFn := getFileProcessWorker(output_file_heap.hash_registry, output_channel, &output_file_heap.size_filter)
	fsobjPool, err := commons.NewWorkerPool(workerFn)

	if err != nil {
		panic(err)
	}

	walker := NewWalker(skip_empty)

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

	walker.SetEntryPoint(start_directory)
	walker.SetDirectoryFilter(get_directory_filter_fn(&user_dirs))
	walker.SetFileCallback(get_file_callback_fn(fsobjPool))
	walker.SetDirectoryCallback(fsobjPool.Wait)

	walker.Walk()

	fsobjPool.Release()

	close(output_channel)

	output_wg.Wait()

	cleaned_heap := output_file_heap.filter_heap(commons.Equal, &shared_registry)
	display_duplicate_file_info(cleaned_heap)

	ui.Close()

	if profile {
		profiler.Collect()
	}
}
