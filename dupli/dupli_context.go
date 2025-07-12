package main

import (
	"fmt"
	"sync"

	"archive-tools-monorepo/commons"
	datastructures "archive-tools-monorepo/dataStructures"
)

type DupliContext struct {
	heap         *datastructures.Heap[commons.File]
	hashRegistry *datastructures.Flyweight[string]
	sizeFilter   sync.Map
}

type DupliContextFunction func(*DupliContext) error

// TODO:
// - make sizeFilter optional
// - convert the function running on heap to methods on context

func WithExistingHeap(heap *datastructures.Heap[commons.File]) DupliContextFunction {
	return func(dc *DupliContext) error {
		dc.heap = heap
		return nil
	}
}

func WithExistingRegistry(registry *datastructures.Flyweight[string]) DupliContextFunction {
	return func(dc *DupliContext) error {
		dc.hashRegistry = registry
		return nil
	}
}

func WithNewHeap(sortFn datastructures.HeapCompareFn[commons.File]) DupliContextFunction {
	newHeap, err := datastructures.NewHeap(
		datastructures.WithComapreFn(sortFn),
		datastructures.WithStartSize[commons.File](1000),
	)

	return func(dc *DupliContext) error {
		if err != nil {
			return fmt.Errorf("error while allocating Heap: \n%w", err)
		}

		dc.heap = newHeap
		return nil
	}
}

func defaultContext() DupliContext {
	return DupliContext{
		heap:         nil,
		hashRegistry: nil,
		sizeFilter:   sync.Map{},
	}
}

func newDupliContext(optsFn ...DupliContextFunction) (*DupliContext, error) {
	var err error
	dupliContext := defaultContext()

	for _, fn := range optsFn {
		err = fn(&dupliContext)
		if err != nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return &dupliContext, nil
}

func (dupliCtx *DupliContext) Display() error {
	var lastSeen commons.File
	var current commons.File
	var areEqual bool
	var err error

	isDuplicate := false

	if !dupliCtx.heap.Empty() {
		current, err = dupliCtx.heap.Pop()
		if err != nil {
			return fmt.Errorf("%w", err)
		}
	}

	for !dupliCtx.heap.Empty() {
		lastSeen = current
		current, err = dupliCtx.heap.Pop()
		if err != nil {
			return fmt.Errorf("%w", err)
		}

		areEqual = commons.WeakFileEqulity(&current, &lastSeen)

		if areEqual {
			ui.Println("file: %s", &lastSeen)
		} else if isDuplicate {
			ui.Println("file: %s", &lastSeen)
		}

		isDuplicate = areEqual
	}

	if isDuplicate {
		ui.Println("file: %s", &lastSeen)
	}

	return nil
}
