package ds

import (
	"sync"
)

type Stack[T any] struct {
	mutex sync.Mutex
	items []T
}

func Is_stack_empty[T any](stack *Stack[T]) bool {
	return len(stack.items) == 0
}

func Push_into_stack[T any](stack *Stack[T], data T) {
	stack.mutex.Lock()
	stack.items = append(stack.items, data)
	stack.mutex.Unlock()
}
    
func Pop_from_stack[T any](stack *Stack[T]) T {
	var item T
	stack.mutex.Lock()
	if len(stack.items) != 0 {
		item = stack.items[len(stack.items)-1]
		stack.items = stack.items[:len(stack.items)-1]
	}
	stack.mutex.Unlock()
	return item
}
    