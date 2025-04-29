package ds

import (
	"sync"
)

type Heap[T any] struct {
	mutex sync.Mutex
	items []T
	custom_is_lower_fn func(T, T)bool
}

func Is_heap_empty[T any](heap *Heap[T]) bool {
	return len(heap.items) == 0
}

func get_left_node_index(index int) int {
	return (index * 2) + 1
}

func get_right_node_index(index int) int {
	return (index * 2) + 2
}

func get_parent_node_index(index int) int {
	if index == 0 {
		return -1
	}

	return (index - 1) / 2
}

func Push_into_heap[T any](heap *Heap[T], data T) {
	heap.mutex.Lock()

	heap.items = append(heap.items, data)
	heapify(heap)

	heap.mutex.Unlock()
}

func heapify[T any](heap *Heap[T]) {
	current_index := len(heap.items) - 1
	parent := get_parent_node_index(current_index)
	var swap_variable T

	for current_index > 0 && !heap.custom_is_lower_fn(heap.items[current_index], heap.items[parent]) {
		swap_variable = heap.items[parent]
		heap.items[parent] = heap.items[current_index]
		heap.items[current_index] = swap_variable

		current_index = parent
		if current_index > 0 {
			parent = get_parent_node_index(current_index)
		}
	}
}

/*    
func Pop_from_heap[T any](heap *Heap[T]) T {
	var item T
	heap.mutex.Lock()
	if len(heap.items) != 0 {
		item = heap.items[len(heap.items)-1]
		heap.items = heap.items[:len(heap.items)-1]
	}
	heap.mutex.Unlock()
	return item
}
*/