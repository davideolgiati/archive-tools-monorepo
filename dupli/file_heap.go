package main

import (
	"fmt"
	"sync"

	"archive-tools-monorepo/commons"
	datastructures "archive-tools-monorepo/dataStructures"
)

func refineFile(file commons.File, fileChannel chan<- commons.File, flyweight *datastructures.Flyweight[string]) error {
	if file.Hash.Value() != "" {
		fileChannel <- file
		return nil
	}

	hash, err := commons.GetSHA1HashFromPath(file.Name)
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

func (dupliCtx *DupliContext) filterHeap(
	filterFunction func(*commons.File, *commons.File) bool,
	registry *datastructures.Flyweight[string],
) *DupliContext {
	var current commons.File
	var last commons.File

	output, err := newDupliContext(
		WithNewHeap(commons.StrongFileCompare),
		WithExistingRegistry(registry),
	)
	if err != nil {
		panic(err)
	}

	output.hashRegistry = dupliCtx.hashRegistry

	total := float64(dupliCtx.heap.Size())
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

	if !dupliCtx.heap.Empty() {
		current, err = dupliCtx.heap.Pop()
		if err != nil {
			panic(err)
		}

		processed += 1.0
	}

	for !dupliCtx.heap.Empty() {
		last = current
		current, err = dupliCtx.heap.Pop()
		if err != nil {
			panic(err)
		}

		processed += 1.0

		switch {
		case filterFunction(&current, &last):
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
