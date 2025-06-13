package dataStructures

import (
	"fmt"
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
	start_size := len(stack.items)

	stack.items = append(stack.items, data)

	if len(stack.items) != start_size+1 {
		panic("Error while stacking")
	}
}

func (stack *Stack[T]) Pop() T {
	stack.mutex.Lock()
	defer stack.mutex.Unlock()

	start_size := len(stack.items)

	var item T
	if len(stack.items) != 0 {
		item = stack.items[len(stack.items)-1]
		stack.items = stack.items[:len(stack.items)-1]

		if len(stack.items) != start_size-1 {
			panic(fmt.Sprintf("Error while popping - popping, start_size: %d, end_size: %d", start_size, len(stack.items)))
		}
	}

	if start_size == 0 && len(stack.items) != start_size {
		panic(fmt.Sprintf("Error while popping - noop, start_size: %d, end_size: %d", start_size, len(stack.items)))
	}

	return item
}
