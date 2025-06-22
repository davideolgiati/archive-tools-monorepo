package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/dataStructures"
	"sync"
)

type FileHeap struct {
	heap         *dataStructures.Heap[commons.File]
	hashRegistry *dataStructures.Flyweight[string]
	sizeFilter   sync.Map
}

func newFileHeap(sortFunction func(*commons.File, *commons.File) bool, registry *dataStructures.Flyweight[string]) (*FileHeap, error) {
	fileHeap := FileHeap{}

	newHeap, err := dataStructures.NewHeap(sortFunction)

	if err != nil {
		return nil, err
	}

	fileHeap.heap = newHeap

	fileHeap.hashRegistry = registry

	return &fileHeap, nil
}

func refineAndPushFileInHeap(file commons.File, file_chan chan<- commons.File, flyweight *dataStructures.Flyweight[string]) {
	if file.Hash.Value() == "" {
		file_chan <- file
		return
	}

	hash, err := commons.CalculateHash(file.Name)

	if err != nil {
		return
		//panic(err)
	}

	file.Hash, err = flyweight.Instance(hash)

	if err != nil {
		return
		//panic(err)
	}

	file_chan <- file
}

func getFileHashGoruotine(file_chan chan<- commons.File, flyweight *dataStructures.Flyweight[string]) func(commons.File) {
	return func(obj commons.File) {
		refineAndPushFileInHeap(obj, file_chan, flyweight)
	}
}

func consumeFromFileChannel(channel chan commons.File, waitgroup *sync.WaitGroup, heap *dataStructures.Heap[commons.File]) {
	defer waitgroup.Done()
	for data := range channel {
		heap.Push(data)
	}
}

func (fh *FileHeap) filterHeap(filterFunction func(commons.File, commons.File) bool, registry *dataStructures.Flyweight[string]) *FileHeap {
	var current commons.File
	var last commons.File

	output, err := newFileHeap(commons.HashDescending, registry)

	if err != nil {
		panic(err)
	}

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
	go consumeFromFileChannel(fileChannel, &outputWaitgroup, output.heap)

	ui.AddNewNamedLine("cleanup-stage", "Removing unique entries %s ... %.1f %%")

	if !fh.heap.Empty() {
		current, err = fh.heap.Pop()

		if err != nil {
			panic(err)
		}

		processed += 1.0
	}

	for !fh.heap.Empty() {
		last = current
		current, err = fh.heap.Pop()

		if err != nil {
			panic(err)
		}

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
	var err error

	isDuplicate := false

	if !fh.heap.Empty() {
		current, err = fh.heap.Pop()
		if err != nil {
			panic(err)
		}
	}

	for !fh.heap.Empty() {
		lastSeen = current
		current, err = fh.heap.Pop()

		if err != nil {
			panic(err)
		}

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
