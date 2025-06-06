package ds

type Queue[T comparable] struct {
	queue []T
}

func (queue *Queue[T]) Init() {
	queue.queue = make([]T, 0)
}

func (queue *Queue[T]) Push(value T) {
	queue.queue = append(queue.queue, value)
}

func (queue *Queue[T]) Pop() T {
	if len(queue.queue) == 0 {
		panic("Pop on empty list")
	}

	current := queue.queue[0]
	queue.queue = queue.queue[1:]

	return current
}

func (queue *Queue[T]) Empty() bool {
	return len(queue.queue) == 0
}
