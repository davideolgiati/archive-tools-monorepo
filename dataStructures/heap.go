package dataStructures

import (
	"math"
	"sync"
)

type Heap[T any] struct {
	items              []T
	tail               int
	size               int
	custom_is_lower_fn func(T, T) bool
	mutex              sync.Mutex
}

func NewHeap[T any](sortFunction func(T, T) bool) *Heap[T] {
	heap := Heap[T]{}

	if sortFunction == nil {
		panic("Provided function is a nil pointer")
	}

	heap.custom_is_lower_fn = sortFunction
	heap.items = make([]T, 0)
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
	if heap.custom_is_lower_fn == nil {
		panic("comapre function not set!")
	}

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	if heap.tail == len(heap.items) {
		heap.resize()
	}

	heap.items[heap.tail] = data
	heap.tail++
	heap.size++

	if heap.size > 1 {
		heap.heapifyBottomUp()
	}
}

func (heap *Heap[T]) Pop() T {
	if heap.custom_is_lower_fn == nil {
		panic("comapre function not set!")
	}

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	var item T

	if heap.size != 0 {
		item = heap.items[0]
		heap.tail--
		heap.size--
		heap.items[0] = heap.items[heap.tail]

		if heap.size != 0 {
			heap.heapifyTopDown()
		}
	}

	return item
}

func (heap *Heap[T]) Peak() *T {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	var item *T

	if heap.size != 0 {
		item = &heap.items[0]
	}

	if heap.size != 0 && item == nil {
		panic("pointer to heap.items[0] is nil")
	}

	return item
}

func getLeftNode(index *int) int {
	if index == nil {
		panic("provide index pointer is nil")
	}

	if *index < 0 {
		panic("provide index is < 0")
	}

	return (*index * 2) + 1
}

func getRightNode(index *int) int {
	if index == nil {
		panic("provide index pointer is nil")
	}

	if *index < 0 {
		panic("provide index is < 0")
	}

	return (*index * 2) + 2
}

func getParent(index *int) int {
	if index == nil {
		panic("provide index pointer is nil")
	}

	if *index <= 0 {
		panic("provide index is <= 0")
	}

	return (*index - 1) / 2
}

func (heap *Heap[T]) heapifyBottomUp() {
	current_index := heap.tail - 1
	parent := getParent(&current_index)

	for current_index > 0 && heap.custom_is_lower_fn(heap.items[current_index], heap.items[parent]) {
		heap.items[parent], heap.items[current_index] = heap.items[current_index], heap.items[parent]

		current_index = parent
		if current_index > 0 {
			parent = getParent(&current_index)
		}

		if current_index < 0 {
			panic("current_index is not positive")
		}

		if parent < 0 {
			panic("parent is not positive")
		}

		if parent > current_index {
			panic("parent is higher than current index")
		}
	}
}

func (heap *Heap[T]) getSmallestChild(current *int) int {
	if heap == nil {
		panic("heap pointer is nil")
	}

	if current == nil {
		panic("current pointer is nil")
	}

	if *current >= heap.tail {
		panic("current is beyond heap scope")
	}

	left := getLeftNode(current)
	right := getRightNode(current)

	if left == right {
		panic("left is right")
	}

	if left >= heap.tail || right > heap.tail {
		return heap.tail
	}

	if right == heap.tail {
		return left
	}

	if heap.custom_is_lower_fn(heap.items[left], heap.items[right]) {
		return left
	} else {
		return right
	}
}

func (heap *Heap[T]) heapifyTopDown() {
	current_index := 0
	candidate := heap.getSmallestChild(&current_index)

	for candidate < heap.tail && heap.custom_is_lower_fn(heap.items[candidate], heap.items[current_index]) {
		heap.items[candidate], heap.items[current_index] = heap.items[current_index], heap.items[candidate]

		current_index = candidate
		candidate = heap.getSmallestChild(&candidate)

		if current_index < 0 {
			panic("current_index is not positive")
		}

		if candidate < 0 {
			panic("candidate is not positive")
		}

		if candidate > heap.tail {
			panic("candidate is beyond heap scope")
		}

		if current_index > heap.tail {
			panic("candidate is beyond heap scope")
		}

		if candidate < current_index {
			panic("candidate is lower than current index")
		}
	}
}

func (heap *Heap[T]) resize() {
	newSize := math.Pow(float64(len(heap.items)), 2) + 1
	newItems := make([]T, uint(newSize))

	copy(newItems[:len(heap.items)], heap.items)

	heap.items = newItems
}
