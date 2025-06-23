package datastructures

import (
	"fmt"
	"math"
	"sync"
)

type Heap[T any] struct {
	items       []*T
	tail        int
	size        int
	minFunction func(*T, *T) bool
	mutex       sync.Mutex
}

func NewHeap[T any](sortFunction func(*T, *T) bool) (*Heap[T], error) {
	heap := Heap[T]{}

	if sortFunction == nil {
		return nil, fmt.Errorf("provided function is a nil pointer")
	}

	heap.minFunction = sortFunction
	heap.items = make([]*T, 0)
	heap.tail = 0
	heap.size = 0

	return &heap, nil
}

func (heap *Heap[T]) Empty() bool {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	return heap.size == 0
}

func (heap *Heap[T]) Size() int {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	return heap.size
}

func (heap *Heap[T]) Push(data T) error {
	if heap == nil || heap.minFunction == nil {
		return fmt.Errorf("comapre function not set")
	}

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	if heap.tail == len(heap.items) {
		heap.resize()
	}

	heap.items[heap.tail] = &data
	heap.tail++
	heap.size++

	heap.heapifyBottomUp()

	return nil
}

func (heap *Heap[T]) Pop() (T, error) {
	var item T

	if heap == nil || heap.minFunction == nil {
		return item, fmt.Errorf("comapre function not set")
	}

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	if heap.size != 0 {
		item = *heap.items[0]
		heap.tail--
		heap.size--

		if heap.size != 0 {
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
	if heap.size != 0 {
		item = heap.items[0]
	}

	return item
}

func (heap *Heap[T]) heapifyBottomUp() {
	currentIndex := heap.tail - 1
	parent := (heap.tail - 2) / 2

	for currentIndex > 0 && heap.minFunction(heap.items[currentIndex], heap.items[parent]) {
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
	case left+1 == heap.tail || heap.minFunction(heap.items[left], heap.items[left+1]):
		return left
	default:
		return left + 1
	}
}

func (heap *Heap[T]) heapifyTopDown() {
	currentIndex := 0
	candidate := heap.getSmallestChild(&currentIndex)

	for candidate < heap.tail && heap.minFunction(heap.items[candidate], heap.items[currentIndex]) {
		heap.items[candidate], heap.items[currentIndex] = heap.items[currentIndex], heap.items[candidate]
		currentIndex = candidate

		if currentIndex*2 > heap.tail {
			break
		}

		candidate = heap.getSmallestChild(&candidate)
	}
}

func (heap *Heap[T]) resize() {
	newSize := math.Pow(float64(len(heap.items)), 2) + 1
	newItems := make([]*T, uint(newSize))

	copy(newItems[:len(heap.items)], heap.items)

	heap.items = newItems
}
