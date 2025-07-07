package main

import (
	"fmt"
	"sync"

	"archive-tools-monorepo/commons"
	datastructures "archive-tools-monorepo/dataStructures"
)

type FileHeap struct {
	heap         *datastructures.Heap[commons.File]
	hashRegistry *datastructures.Flyweight[string]
	sizeFilter   sync.Map
}

// TODO:
// - rename FileHeap into DupliContext
// - apply creational pattern to DupliContext like in Heap
// - add WithNewHeap, WithExixtingHeap, WithNewHashRegistry, WithExixtingRegistry
// - make sizeFilter optional
// - convert the function running on heap to methods on context

func newFileHeap(
	sortFunction datastructures.HeapCompareFn[commons.File],
	registry *datastructures.Flyweight[string],
) (*FileHeap, error) {
	newHeap, err := datastructures.NewHeap(
		datastructures.WithComapreFn(sortFunction),
		datastructures.WithStartSize[commons.File](1000),
	)
	if err != nil {
		return nil, fmt.Errorf("error while allocating Heap: \n%w", err)
	}

	fileHeap := FileHeap{
		heap:         newHeap,
		hashRegistry: registry,
		sizeFilter:   sync.Map{},
	}

	return &fileHeap, nil
}

func refineFile(file commons.File, fileChannel chan<- commons.File, flyweight *datastructures.Flyweight[string]) error {
	if file.Hash.Value() != "" {
		fileChannel <- file
		return nil
	}

	hash, err := commons.CalculateHash(file.Name)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	file.Hash, err = flyweight.Instance(hash)
	if err != nil {
		return fmt.Errorf("%w", err)
	}

	fileChannel <- file
	return nil
}

func getFileHashGoruotine(
	fileChannel chan<- commons.File,
	flyweight *datastructures.Flyweight[string],
) func(commons.File) error {
	return func(obj commons.File) error {
		return refineFile(obj, fileChannel, flyweight)
	}
}

func consumeFromFileChannel(
	channel chan commons.File,
	waitgroup *sync.WaitGroup,
	heap *datastructures.Heap[commons.File],
) {
	var err error
	defer waitgroup.Done()
	for data := range channel {
		err = heap.Push(data)
		if err != nil {
			continue // TODO: this need to be fixed
		}
	}
}

func (fh *FileHeap) filterHeap(
	filterFunction func(commons.File, commons.File) bool,
	registry *datastructures.Flyweight[string],
) *FileHeap {
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
