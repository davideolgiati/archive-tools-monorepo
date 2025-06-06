package ds

type node[T any] struct {
	value T
	priority uint
}

type PriorityQueue[T comparable] struct {
	queue Heap[node[T]]
}

func (queue *PriorityQueue[T]) Init() {
	queue.queue = Heap[node[T]]{}
	queue.queue.custom_is_lower_fn = func(n1, n2 node[T]) bool {
		return n1.priority < n2.priority
	}
}

func (queue *PriorityQueue[T]) Push(value T, priority uint) {
	new_node := node[T]{
		value: value,
		priority: priority,
	}

	queue.queue.Push(new_node)
}

func (queue *PriorityQueue[T]) Pop() (uint, T) {
	if queue.queue.Empty() {
		panic("Pop on empty list")
	}

	current_node := queue.queue.Pop()

	return current_node.priority, current_node.value
}

func (queue *PriorityQueue[T]) Empty() bool {
	return queue.queue.Empty()
}
