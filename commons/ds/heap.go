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

func (heap *Heap[T]) Compare_fn(custom_is_lower_fn func(T, T) bool) {
	heap.custom_is_lower_fn = custom_is_lower_fn
}

func (heap *Heap[T]) Empty() bool {
	return len(heap.items) == 0
}

func (heap *Heap[T]) Size() int {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	return len(heap.items)
}

func (heap *Heap[T]) Push(data T) {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()

	start_size := len(heap.items)

	heap.items = append(heap.items, data)
	heapify_bottom_up(heap)

	if len(heap.items) != start_size + 1 {
		panic(fmt.Sprintf("wrong heap size, expected %d, got %d", start_size + 1, len(heap.items)))
	}
}

func (heap *Heap[T]) Pop() T {
	heap.mutex.Lock()
	defer heap.mutex.Unlock()
	
	var item T

	start_size := len(heap.items)

	if len(heap.items) != 0 {
		item = heap.items[0]
		heap.items[0] = heap.items[len(heap.items)-1]
		heap.items = heap.items[:len(heap.items)-1]

		if(!heap.Empty()) {
			heapify_top_down(heap)
		}
	}
	
	if len(heap.items) != start_size - 1 {
		panic(fmt.Sprintf("wrong heap size, expected %d, got %d", start_size - 1, len(heap.items)))
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
	
	return item
}

func get_left_node_index(index *int) int {
	return (*index * 2) + 1
}

func get_right_node_index(index *int) int {
	return (*index * 2) + 2
}

func get_parent_node_index(index *int) int {
	if *index == 0 {
		return -1
	}

	return (*index - 1) / 2
}

func heapify_bottom_up[T any](heap *Heap[T]) {
	current_index := len(heap.items) - 1
	parent := get_parent_node_index(&current_index)
	var swap_variable T

	for current_index > 0 && heap.custom_is_lower_fn(heap.items[current_index], heap.items[parent]) {
		swap_variable = heap.items[parent]
		heap.items[parent] = heap.items[current_index]
		heap.items[current_index] = swap_variable

		current_index = parent
		
		if current_index > 0 {
			parent = get_parent_node_index(&current_index)
		}
	}
}

func get_next_candidate[T any](heap *Heap[T], current *int) int {
	left := get_left_node_index(current)
	right := get_right_node_index(current)

	if left >= len(heap.items) {
		return len(heap.items) - 1
	}

	if right >= len(heap.items) {
		return left
	}

	if heap.custom_is_lower_fn(heap.items[left], heap.items[right]) {
		return left
	} else {
		return right
	}
}

func heapify_top_down[T any](heap *Heap[T]) {
	current_index := 0
	candidate := get_next_candidate(heap, &current_index)

	var swap_variable T

	for heap.custom_is_lower_fn(heap.items[candidate], heap.items[current_index]) {
		swap_variable = heap.items[candidate]
		heap.items[candidate] = heap.items[current_index]
		heap.items[current_index] = swap_variable

		current_index = candidate
		candidate = get_next_candidate(heap, &candidate)
	}
}