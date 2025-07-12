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

func processOutputChannelData(
	outputChannel chan commons.File,
	outputWg *sync.WaitGroup,
	container *datastructures.Heap[commons.File],
) {
	var err error
	for data := range outputChannel {
		err = container.Push(data)
		if err != nil {
			panic(err)
		}
	}
	outputWg.Done()
}

func main() {
	startDirectory := ""
	ignoredDirUser := ""
	skipEmpty := false
	profile := false
	profiler := commons.Profiler{}

	var fileProcessorPool *commons.WriteOnlyThreadPool[FilesystemObject]

	sharedRegistry := datastructures.Flyweight[string]{}
	outputFileHeap, err := newDupliContext(
		WithNewHeap(commons.StrongFileCompare),
		WithExistingRegistry(&sharedRegistry),
	)
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

	if outputFileHeap == nil {
		panic("error wile creating new file heap object")
	}

	workerFn, err := getFileProcessWorker(outputFileHeap.hashRegistry, outputChannel, &outputFileHeap.sizeFilter)
	if err != nil {
		panic(err)
	}

	fileProcessorPool, err = commons.NewWorkerPool(workerFn)
	if err != nil {
		panic(err)
	}

	walker := NewWalker(skipEmpty)

	if walker == nil {
		panic("error wile creating new file walker object")
	}

	outputWg.Add(1)

	go processOutputChannelData(outputChannel, &outputWg, outputFileHeap.heap)

	walker.SetEntryPoint(startDirectory)
	walker.SetDirectoryFilter(getDirectoryFilter(&userDirectories))
	walker.SetFileCallback(getFileCallback(fileProcessorPool))
	walker.SetDirectoryCallback(fileProcessorPool.Wait)

	walker.Walk()

	fileProcessorPool.Release()

	close(outputChannel)

	outputWg.Wait()

	cleanedHeap := outputFileHeap.filterHeap(commons.StrongFileEquality, &sharedRegistry)
	err = cleanedHeap.Display()

	ui.Close()

	if err != nil {
		panic(err)
	}

	if profile {
		err = profiler.Collect()
		if err != nil {
			panic(err)
		}
	}
}
