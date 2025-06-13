package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"sync"
)

type FileHeap struct {
	heap         ds.Heap[commons.File]
	hashRegistry *ds.Flyweight[string]
	sizeFilter   sync.Map
}

func newFileHeap(sortFunction func(commons.File, commons.File) bool, registry *ds.Flyweight[string]) *FileHeap {
	fileHeap := FileHeap{}

	fileHeap.heap = *ds.NewHeap(sortFunction)
	fileHeap.hashRegistry = registry

	return &fileHeap
}

func refineAndPushFileInHeap(file commons.File, file_chan chan<- commons.File, flyweight *ds.Flyweight[string]) {
	if file.Hash.Value() == "" {
		file_chan <- file
		return
	}

	hash, err := commons.CalculateHash(file.Name)

	if err != nil {
		panic(err)
	}

	file.Hash, err = flyweight.Instance(hash)

	if err != nil {
		panic(err)
	}

	file_chan <- file
}

func getFileHashGoruotine(file_chan chan<- commons.File, flyweight *ds.Flyweight[string]) func(commons.File) {
	return func(obj commons.File) {
		refineAndPushFileInHeap(obj, file_chan, flyweight)
	}
}

func consumeFromFileChannel(channel chan commons.File, waitgroup *sync.WaitGroup, heap *ds.Heap[commons.File]) {
	defer waitgroup.Done()
	for data := range channel {
		heap.Push(data)
	}
}

func (fh *FileHeap) filterHeap(filterFunction func(commons.File, commons.File) bool, registry *ds.Flyweight[string]) *FileHeap {
	var current commons.File
	var last commons.File

	output := newFileHeap(commons.HashDescending, registry)
	output.hashRegistry = fh.hashRegistry

	total := float64(fh.heap.Size())
	processed := 0.0

	duplicateFlag := false

	fileChannel := make(chan commons.File)
	targetFunction := getFileHashGoruotine(fileChannel, output.hashRegistry)
	fileThreadPool, err := commons.NewWorkerPool(targetFunction)

	if err != nil {
		panic(err)
	}

	outputWaitgroup := sync.WaitGroup{}
	outputWaitgroup.Add(1)
	go consumeFromFileChannel(fileChannel, &outputWaitgroup, &output.heap)

	ui.AddNewNamedLine("cleanup-stage", "Removing unique entries %s ... %.1f %%")

	if !fh.heap.Empty() {
		current = fh.heap.Pop()
		processed += 1.0
	}

	for !fh.heap.Empty() {
		last = current
		current = fh.heap.Pop()
		processed += 1.0

		if filterFunction(current, last) {
			duplicateFlag = true
			fileThreadPool.Submit(last)
		} else if duplicateFlag {
			duplicateFlag = false
			fileThreadPool.Submit(last)
		} else {
			duplicateFlag = false
		}

		ui.UpdateNamedLine("cleanup-stage", "cleanup-stage", (processed/total)*100)
	}

	if duplicateFlag {
		fileThreadPool.Submit(current)
	}

	fileThreadPool.Release()
	close(fileChannel)
	outputWaitgroup.Wait()

	return output
}

func (fh *FileHeap) display_duplicate_file_info() {
	var lastSeen commons.File
	var current commons.File
	var areEqual bool

	isDuplicate := false

	if !fh.heap.Empty() {
		current = fh.heap.Pop()
	}

	for !fh.heap.Empty() {
		lastSeen = current
		current = fh.heap.Pop()
		areEqual = commons.EqualByHash(current, lastSeen)

		if areEqual {
			ui.Println("file: %s", lastSeen)
		} else {
			if isDuplicate {
				ui.Println("file: %s", lastSeen)
			}
		}

		isDuplicate = areEqual

	}

	if isDuplicate {
		ui.Println("file: %s", lastSeen)
	}
}
