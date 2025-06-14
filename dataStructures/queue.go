package dataStructures

import "fmt"

type Queue[T comparable] struct {
	queue []T
	head  int
	tail  int
	size  int
}

func (queue *Queue[T]) Init() {
	queue.queue = make([]T, 30)
	queue.head = 0
	queue.tail = 0
	queue.size = 0
}

func (queue *Queue[T]) Push(value T) {
	if queue.size == len(queue.queue) {
		queue.resize()
	}

	queue.queue[queue.tail] = value
	queue.tail = (queue.tail + 1) % len(queue.queue)
	queue.size++
}

func (queue *Queue[T]) Pop() (T, error) {
	var value T

	if queue.size == 0 {
		return value, fmt.Errorf("Pop on empty queue")
	}

	value = queue.queue[queue.head]
	queue.head = (queue.head + 1) % len(queue.queue)
	queue.size--

	return value, nil
}

func (queue *Queue[T]) Empty() bool {
	return queue.size == 0
}

func (queue *Queue[T]) Size() int {
	return queue.size
}

func (queue *Queue[T]) resize() {
	newQueue := make([]T, len(queue.queue)*2)

	for i := 0; i < queue.size; i++ {
		newQueue[i] = queue.queue[(queue.head+i)%len(queue.queue)]
	}

	queue.queue = newQueue
	queue.head = 0
	queue.tail = queue.size
}
