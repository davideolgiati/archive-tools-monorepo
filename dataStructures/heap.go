package datastructures

import (
	"errors"
	"sync"
)

type (
	HeapCompareFn[T any] func(*T, *T) bool
	OptsFn[T any]        func(*Opts[T])
)

type Opts[T any] struct {
	comapreFn HeapCompareFn[T]
	size      int
}

type Heap[T any] struct {
	opts         Opts[T]
	items        []*T
	elementCount int
	tail         int
	mutex        sync.Mutex
}

func defaultOpts[T any]() Opts[T] {
	return Opts[T]{
		size:      0,
		comapreFn: nil,
	}
}

func WithComapreFn[T any](fn HeapCompareFn[T]) OptsFn[T] {
	return func(o *Opts[T]) {
		o.comapreFn = fn
	}
}

func WithStartSize[T any](size int) OptsFn[T] {
	return func(o *Opts[T]) {
		o.size = size
	}
}

func NewHeap[T any](optsFunctions ...OptsFn[T]) (*Heap[T], error) {
	baseOpts := defaultOpts[T]()

	for _, fn := range optsFunctions {
		fn(&baseOpts)
	}

	if baseOpts.comapreFn == nil {
		return nil, errors.New("provided function is a nil pointer")
	}

	heap := Heap[T]{
		opts:         baseOpts,
		elementCount: 0,
		items:        make([]*T, baseOpts.size),
		tail:         0,
		mutex:        sync.Mutex{},
	}

	return &heap, nil
}

func (heap *Heap[T]) Empty() bool {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	return heap.elementCount == 0
}

func (heap *Heap[T]) Size() int {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	return heap.elementCount
}

func (heap *Heap[T]) Push(data T) error {
	if heap == nil || heap.opts.comapreFn == nil {
		return errors.New("comapre function not set")
	}

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	if heap.tail == heap.opts.size {
		heap.resize()
	}

	heap.items[heap.tail] = &data
	heap.tail++
	heap.elementCount++

	heap.heapifyBottomUp()

	return nil
}

func (heap *Heap[T]) Pop() (T, error) {
	var item T

	if heap == nil || heap.opts.comapreFn == nil {
		return item, errors.New("comapre function not set")
	}

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	if heap.elementCount != 0 {
		item = *heap.items[0]
		heap.tail--
		heap.elementCount--

		if heap.elementCount != 0 {
			heap.items[0] = heap.items[heap.tail]
			heap.heapifyTopDown()
		}
	}

	return item, nil
}

func (heap *Heap[T]) Peak() *T {
	var item *T

	heap.mutex.Lock()
	defer heap.mutex.Unlock()
	if heap.elementCount != 0 {
		item = heap.items[0]
	}

	return item
}

func (heap *Heap[T]) heapifyBottomUp() {
	currentIndex := heap.tail - 1
	parent := (heap.tail - 2) / 2

	for currentIndex > 0 && heap.opts.comapreFn(heap.items[currentIndex], heap.items[parent]) {
		heap.items[parent], heap.items[currentIndex] = heap.items[currentIndex], heap.items[parent]

		currentIndex = parent
		parent = (currentIndex - 1) / 2
	}
}

func (heap *Heap[T]) getSmallestChild(current *int) int {
	left := ((*current) * 2) + 1

	switch {
	case left >= heap.tail || left+1 > heap.tail:
		return heap.tail
	case left+1 == heap.tail || heap.opts.comapreFn(heap.items[left], heap.items[left+1]):
		return left
	default:
		return left + 1
	}
}

func (heap *Heap[T]) heapifyTopDown() {
	currentIndex := 0
	candidate := heap.getSmallestChild(&currentIndex)

	for candidate < heap.tail && heap.opts.comapreFn(heap.items[candidate], heap.items[currentIndex]) {
		heap.items[candidate], heap.items[currentIndex] = heap.items[currentIndex], heap.items[candidate]
		currentIndex = candidate

		if currentIndex*2 > heap.tail {
			break
		}

		candidate = heap.getSmallestChild(&candidate)
	}
}

func (heap *Heap[T]) resize() {
	newSize := heap.opts.size*2 + 1
	newItems := make([]*T, uint(newSize))

	copy(newItems[:heap.opts.size], heap.items)

	heap.opts.size = newSize

	heap.items = newItems
}
