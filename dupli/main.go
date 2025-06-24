package main

import (
	_ "embed"
	"flag"
	"strings"
	"sync"

	"archive-tools-monorepo/commons"
	datastructures "archive-tools-monorepo/dataStructures"
)

//go:embed logo.txt
var logo string

//go:embed semver.txt
var version string

//go:embed buildts.txt
var buildts string

var ui = commons.NewUI()

func filter[T comparable](input []T, filterValue T) []T {
	output := make([]T, 0)

	for _, str := range input {
		if str == filterValue {
			continue
		}

		output = append(output, str)
	}

	return output
}

func main() {
	startDirectory := ""
	ignoredDirUser := ""
	skipEmpty := false
	profile := false
	profiler := commons.Profiler{}

	var fileProcessorPool *commons.WriteOnlyThreadPool[File]

	sharedRegistry := datastructures.Flyweight[string]{}
	outputFileHheap, err := newFileHeap(commons.HashDescending, &sharedRegistry)
	if err != nil {
		panic(err)
	}

	outputChannel := make(chan commons.File)
	outputWg := sync.WaitGroup{}

	flag.StringVar(&startDirectory, "dir", "", "Scan starting point  directory")
	flag.StringVar(&ignoredDirUser, "skip_dirs", "", "Skip user defined directories during scan (separated by comma)")
	flag.BoolVar(&skipEmpty, "no_empty", false, "Skip empty files during scan")
	flag.BoolVar(&profile, "profile", false, "Profile program performances")

	flag.Parse()

	if profile {
		ui.ToggleSilence()
		profiler.Start()
	}

	userDirectories := filter(strings.Split(ignoredDirUser, ","), "")

	ui.Println("%s", logo)
	ui.Println("Running version: %s", version)
	ui.Println("Build timestamp: %s", buildts)

	if outputFileHheap == nil {
		panic("error wile creating new file heap object")
	}

	workerFn := getFileProcessWorker(outputFileHheap.hashRegistry, outputChannel, &outputFileHheap.sizeFilter)
	fileProcessorPool, err = commons.NewWorkerPool(workerFn)
	if err != nil {
		panic(err)
	}

	walker := NewWalker(skipEmpty)

	if walker == nil {
		panic("error wile creating new file walker object")
	}

	outputWg.Add(1)

	go func() {
		var err error
		for data := range outputChannel {
			err = outputFileHheap.heap.Push(data)
			if err != nil {
				panic(err)
			}
		}
		outputWg.Done()
	}()

	walker.SetEntryPoint(startDirectory)
	walker.SetDirectoryFilter(getDirectoryFilter(&userDirectories))
	walker.SetFileCallback(getFileCallback(fileProcessorPool))
	walker.SetDirectoryCallback(fileProcessorPool.Wait)

	walker.Walk()

	fileProcessorPool.Release()

	close(outputChannel)

	outputWg.Wait()

	cleanedHeap := outputFileHheap.filterHeap(commons.Equal, &sharedRegistry)
	cleanedHeap.displayDuplicateFileInfo()

	ui.Close()

	if profile {
		profiler.Collect()
	}
}
