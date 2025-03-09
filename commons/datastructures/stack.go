package datastructures

import (
	"fmt"
	"sync"
)

type Stack struct {
	mutex sync.Mutex
	items []string
}

func Is_stack_empty(stack *Stack) bool {
	return len(stack.items) == 0
}

func Push_into_stack(stack *Stack, data string) {
	stack.mutex.Lock()
	stack.items = append(stack.items, data)
	stack.mutex.Unlock()
}
    
func Pop_from_stack(stack *Stack) {
	if Is_stack_empty(stack) {
		return
	}

	stack.mutex.Lock()
	stack.items = stack.items[:len(stack.items)-1]
	stack.mutex.Unlock()
}

func Get_stack_top_element(stack *Stack) (string, error) {
	if Is_stack_empty(stack) {
	    return "", fmt.Errorf("stack is empty")
	}

	stack.mutex.Lock()
	data := stack.items[len(stack.items)-1]
	stack.mutex.Unlock()

	return data, nil
}
    