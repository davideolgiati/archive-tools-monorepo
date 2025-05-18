package ds

import (
	"sync"
)

type Stack[T any] struct {
	items []T
	mutex sync.Mutex
}

func (stack *Stack[T]) Empty() bool {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	
	return len(stack.items) == 0
}

func (stack *Stack[T]) Push(data T) {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	stack.items = append(stack.items, data)
}
    
func (stack *Stack[T]) Pop() T {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()
	
	var item T
	if len(stack.items) != 0 {
		item = stack.items[len(stack.items)-1]
		stack.items = stack.items[:len(stack.items)-1]
	}

	return item
}
    