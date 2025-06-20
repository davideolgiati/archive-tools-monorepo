package dataStructures

import (
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

func NewHeap[T any](sortFunction func(*T, *T) bool) *Heap[T] {
	heap := Heap[T]{}

	if sortFunction == nil {
		panic("Provided function is a nil pointer")
	}

	heap.minFunction = sortFunction
	heap.items = make([]*T, 0)
	heap.tail = 0
	heap.size = 0

	return &heap
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

func (heap *Heap[T]) Push(data T) {
	if heap.minFunction == nil {
		panic("comapre function not set!")
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
}

func (heap *Heap[T]) Pop() T {
	if heap.minFunction == nil {
		panic("comapre function not set!")
	}

	var item T

	if heap.size != 0 {
		heap.mutex.Lock()
		defer heap.mutex.Unlock()

		item = *heap.items[0]
		heap.tail--
		heap.size--

		if heap.size != 0 {
			heap.items[0] = heap.items[heap.tail]
			heap.heapifyTopDown()
		}
	}

	return item
}

func (heap *Heap[T]) Peak() *T {
	var item *T

	if heap.size != 0 {
		heap.mutex.Lock()
		defer heap.mutex.Unlock()

		item = heap.items[0]
	}

	return item
}

func (heap *Heap[T]) heapifyBottomUp() {
	current_index := heap.tail - 1
	parent := (heap.tail - 2) / 2

	for current_index > 0 && heap.minFunction(heap.items[current_index], heap.items[parent]) {
		heap.items[parent], heap.items[current_index] = heap.items[current_index], heap.items[parent]

		current_index = parent
		parent = (current_index - 1) / 2
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
	current_index := 0
	candidate := heap.getSmallestChild(&current_index)

	for candidate < heap.tail && heap.minFunction(heap.items[candidate], heap.items[current_index]) {
		heap.items[candidate], heap.items[current_index] = heap.items[current_index], heap.items[candidate]
		current_index = candidate

		if current_index * 2 > heap.tail{
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
