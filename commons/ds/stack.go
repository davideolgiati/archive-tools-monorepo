package ds

import (
	"sync"
)

type Stack[T any] struct {
	mutex sync.Mutex
	items []T
}

func (stack *Stack[T]) Empty() bool {
	return len(stack.items) == 0
}

func (stack *Stack[T]) Push(data T) {
	stack.mutex.Lock()
	stack.items = append(stack.items, data)
	stack.mutex.Unlock()
}
    
func (stack *Stack[T]) Pop() T {
	var item T
	stack.mutex.Lock()
	if len(stack.items) != 0 {
		item = stack.items[len(stack.items)-1]
		stack.items = stack.items[:len(stack.items)-1]
	}
	stack.mutex.Unlock()
	return item
}
    