package ds

import (
	"sync"
)

type Heap[T any] struct {
	mutex sync.Mutex
	head *HeapNode[T]
}

type HeapNode[T any] struct {
	value T
	father *HeapNode[T]
	left *HeapNode[T]
	right *HeapNode[T]
}

func Is_heap_empty[T any](heap *Heap[T]) bool {
	return heap.head == nil
}

/*
func Push_into_heap[T any](heap *Heap[T], data T) {
	heap.mutex.Lock()
	heap.items = append(heap.items, data)
	heap.mutex.Unlock()
}
    
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