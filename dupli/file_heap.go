package main

import (
	"sync"

	"archive-tools-monorepo/commons"
	datastructures "archive-tools-monorepo/dataStructures"
)

type FileHeap struct {
	heap         *datastructures.Heap[commons.File]
	hashRegistry *datastructures.Flyweight[string]
	sizeFilter   sync.Map
}

func newFileHeap(sortFunction func(*commons.File, *commons.File) bool, registry *datastructures.Flyweight[string]) (*FileHeap, error) {
	fileHeap := FileHeap{}

	newHeap, err := datastructures.NewHeap(sortFunction)
	if err != nil {
		return nil, err
	}

	fileHeap.heap = newHeap

	fileHeap.hashRegistry = registry

	return &fileHeap, nil
}

func refineFile(file commons.File, fileChannel chan<- commons.File, flyweight *datastructures.Flyweight[string]) {
	if file.Hash.Value() == "" {
		fileChannel <- file
		return
	}

	hash, err := commons.CalculateHash(file.Name)
	if err != nil {
		return
	}

	file.Hash, err = flyweight.Instance(hash)
	if err != nil {
		return
	}

	fileChannel <- file
}

func getFileHashGoruotine(fileChannel chan<- commons.File, flyweight *datastructures.Flyweight[string]) func(commons.File) {
	return func(obj commons.File) {
		refineFile(obj, fileChannel, flyweight)
	}
}

func consumeFromFileChannel(channel chan commons.File, waitgroup *sync.WaitGroup, heap *datastructures.Heap[commons.File]) {
	defer waitgroup.Done()
	for data := range channel {
		err := heap.Push(data)
		if err != nil {
			panic(err)
		}
	}
}

func (fh *FileHeap) filterHeap(filterFunction func(commons.File, commons.File) bool, registry *datastructures.Flyweight[string]) *FileHeap {
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

		switch {
		case filterFunction(current, last):
			duplicateFlag = true
			err = fileThreadPool.Submit(last)
		case duplicateFlag:
			duplicateFlag = false
			err = fileThreadPool.Submit(last)
		default:
			duplicateFlag = false
		}

		if err != nil {
			panic(err)
		}

		ui.UpdateNamedLine("cleanup-stage", "cleanup-stage", (processed/total)*100)
	}

	if duplicateFlag {
		err = fileThreadPool.Submit(current)
		if err != nil {
			panic(err)
		}
	}

	fileThreadPool.Release()
	close(fileChannel)
	outputWaitgroup.Wait()

	return output
}

func (fh *FileHeap) displayDuplicateFileInfo() {
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
		} else if isDuplicate {
			ui.Println("file: %s", lastSeen)
		}

		isDuplicate = areEqual
	}

	if isDuplicate {
		ui.Println("file: %s", lastSeen)
	}
}
