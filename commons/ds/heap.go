package ds

import (
	"fmt"
	"sync"
)

type Heap[T any] struct {
	items              []T
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
	
	return &heap
}

func (heap *Heap[T]) Empty() bool {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	return len(heap.items) == 0
}

func (heap *Heap[T]) Size() int {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	return len(heap.items)
}

func (heap *Heap[T]) Push(data T) {
	if heap.custom_is_lower_fn == nil {
		panic("comapre function not set!")
	}

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	start_size := len(heap.items)

	heap.items = append(heap.items, data)

	if len(heap.items) > 1 {
		heap.heapifyBottomUp()
	}

	if len(heap.items) != start_size+1 {
		panic(fmt.Sprintf("wrong heap size, expected %d, got %d", start_size+1, len(heap.items)))
	}
}

func (heap *Heap[T]) Pop() T {
	if heap.custom_is_lower_fn == nil {
		panic("comapre function not set!")
	}

	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	var item T

	start_size := len(heap.items)

	if len(heap.items) != 0 {
		item = heap.items[0]
		heap.items[0] = heap.items[len(heap.items)-1]
		heap.items = heap.items[:len(heap.items)-1]

		if len(heap.items) != 0 {
			heap.heapifyTopDown()
		}
	}

	if start_size != 0 && len(heap.items) != start_size-1 {
		panic(fmt.Sprintf("wrong heap size, expected %d, got %d", start_size-1, len(heap.items)))
	}

	return item
}

func (heap *Heap[T]) Peak() *T {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	var item *T

	if len(heap.items) != 0 {
		item = &heap.items[0]
	}

	if len(heap.items) != 0 && item == nil {
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
	start_size := len(heap.items)
	current_index := len(heap.items) - 1
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

	if len(heap.items) != start_size {
		panic(fmt.Sprintf("dataloss - heapsize start: %d, heapsize end: %d", start_size, len(heap.items)))
	}
}

func (heap *Heap[T]) getSmallestChild(current *int) int {
	if heap == nil {
		panic("heap pointer is nil")
	}

	if current == nil {
		panic("current pointer is nil")
	}

	if *current >= len(heap.items) {
		panic("current is beyond heap scope")
	}

	left := getLeftNode(current)
	right := getRightNode(current)

	if left == right {
		panic("left is right")
	}

	if left >= len(heap.items) || right > len(heap.items) {
		return len(heap.items)
	}

	if right == len(heap.items) {
		return left
	}

	if heap.custom_is_lower_fn(heap.items[left], heap.items[right]) {
		return left
	} else {
		return right
	}
}

func (heap *Heap[T]) heapifyTopDown() {
	start_size := len(heap.items)
	current_index := 0
	candidate := heap.getSmallestChild(&current_index)

	for candidate < len(heap.items) && heap.custom_is_lower_fn(heap.items[candidate], heap.items[current_index]) {
		heap.items[candidate], heap.items[current_index] = heap.items[current_index], heap.items[candidate]

		current_index = candidate
		candidate = heap.getSmallestChild(&candidate)

		if current_index < 0 {
			panic("current_index is not positive")
		}

		if candidate < 0 {
			panic("candidate is not positive")
		}

		if candidate > len(heap.items) {
			panic("candidate is beyond heap scope")
		}


		if current_index > len(heap.items) {
			panic("candidate is beyond heap scope")
		}

		if candidate < current_index {
			panic("candidate is lower than current index")
		}
	}

	if len(heap.items) != start_size {
		panic(fmt.Sprintf("dataloss - heapsize start: %d, heapsize end: %d", start_size, len(heap.items)))
	}
}
