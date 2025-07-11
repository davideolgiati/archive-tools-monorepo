package datastructures

import (
	"errors"
)

type node[T any] struct {
	value T
	next  *node[T]
}

type Queue[T any] struct {
	head *node[T]
	tail *node[T]
	size int
}

func (queue *Queue[T]) Init() {
	queue.head = nil
	queue.tail = nil
	queue.size = 0
}

func (queue *Queue[T]) Push(value T) {
	newNode := node[T]{
		value: value,
		next:  nil,
	}

	if queue.head == nil {
		queue.head = &newNode
	} else {
		queue.tail.next = &newNode
	}

	queue.tail = &newNode
	queue.size++
}

func (queue *Queue[T]) Pop() (T, error) {
	var value T

	switch queue.size {
	case 0:
		return value, errors.New("Pop on empty queue")
	case 1:
		data := queue.tail.value
		queue.head = nil
		queue.tail = nil
		queue.size = 0
		return data, nil
	default:
		data := queue.head.value
		queue.head = queue.head.next
		queue.size--
		return data, nil
	}
}

func (queue *Queue[T]) Peak() *T {
	var value *T

	switch queue.size {
	case 0:
		return value
	default:
		return &queue.head.value
	}
}

func (queue *Queue[T]) Empty() bool {
	return queue.size == 0
}

func (queue *Queue[T]) Size() int {
	return queue.size
}
