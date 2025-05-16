package ds

import (
	"testing"
)

func TestParentLeftRight(t *testing.T) {
	node_index := 1

	expected_parent := 0
	expected_left := 3
	expected_right := 4

	actual_parent := get_parent_node_index(&node_index)
	actual_left := get_left_node_index(&node_index)
	actual_right := get_right_node_index(&node_index)

	if actual_parent != expected_parent {
		t.Errorf(
			"get_parent_node_index(%d) = %d, expected %d",
			node_index, actual_parent, expected_parent,
		)
	}

	if actual_left != expected_left {
		t.Errorf(
			"get_left_node_index(%d) = %d, expected %d",
			node_index, actual_left, expected_left,
		)
	}

	if actual_right != expected_right {
		t.Errorf(
			"get_right_node_index(%d) = %d, expected %d",
			node_index, actual_right, expected_right,
		)
	}
}

func TestParentLeftRightHead(t *testing.T) {
	node_index := 0

	expected_parent := -1
	expected_left := 1
	expected_right := 2

	actual_parent := get_parent_node_index(&node_index)
	actual_left := get_left_node_index(&node_index)
	actual_right := get_right_node_index(&node_index)

	if actual_parent != expected_parent {
		t.Errorf(
			"get_parent_node_index(%d) = %d, expected %d",
			node_index, actual_parent, expected_parent,
		)
	}

	if actual_left != expected_left {
		t.Errorf(
			"get_left_node_index(%d) = %d, expected %d",
			node_index, actual_left, expected_left,
		)
	}

	if actual_right != expected_right {
		t.Errorf(
			"get_right_node_index(%d) = %d, expected %d",
			node_index, actual_right, expected_right,
		)
	}
}

func is_lower(a *int, b *int) bool {
	return *a < *b
}

func TestFullHeapWorkflow(t *testing.T) {
	input_heap := Heap[int]{}
	input_heap.custom_is_lower_fn = is_lower

	values := []int{1, 30, -1, 20, 25}

	input_heap.Push(values[0])

	if len(input_heap.items) != 1 {
		t.Errorf("len(input_heap.items) = %d, expected 1", len(input_heap.items))
	}

	if *input_heap.Peak() != values[0] {
		t.Errorf("input_heap.items[0] = %d, expected 1", *input_heap.Peak())
	}

	input_heap.Push(values[1])

	if len(input_heap.items) != 2 {
		t.Errorf("len(input_heap.items) = %d, expected 1", len(input_heap.items))
	}

	if *input_heap.Peak() != values[0] {
		t.Errorf("input_heap.items[0] = %d, expected 1", *input_heap.Peak())
	}

	input_heap.Push(values[2])

	if len(input_heap.items) != 3 {
		t.Errorf("len(input_heap.items) = %d, expected 1", len(input_heap.items))
	}

	if *input_heap.Peak() != values[2] {
		t.Errorf("input_heap.items[0] = %d, expected 1", *input_heap.Peak())
	}

	input_heap.Push(values[3])

	if len(input_heap.items) != 4 {
		t.Errorf("len(input_heap.items) = %d, expected 1", len(input_heap.items))
	}

	if *input_heap.Peak() != values[2] {
		t.Errorf("input_heap.items[0] = %d, expected 1", *input_heap.Peak())
	}

	input_heap.Push(values[4])

	if len(input_heap.items) != 5 {
		t.Errorf("len(input_heap.items) = %d, expected 1", len(input_heap.items))
	}

	if *input_heap.Peak() != values[2] {
		t.Errorf("input_heap.items[0] = %d, expected 1", *input_heap.Peak())
	}

	expected_array := []int{-1, 1, 20, 25, 30}
	var current_element int

	for i := 0; i < 5; i++ {
		current_element = input_heap.Pop()
		if current_element != expected_array[i] {
			t.Errorf(
				"Pop_from_heap(&input_heap) = %d, expected %d",
				current_element, expected_array[i],
			)
		}
	}

}

func TestHeapifyBottomUpBubbleUp(t *testing.T) {
	// Create a valid min-heap: [1, 5, 10, 20]
	// Then simulate inserting a new element "0" at the end that needs to bubble up.
	h := Heap[int]{
		custom_is_lower_fn: is_lower,
		items:             []int{1, 5, 10, 20},
	}
	h.items = append(h.items, 0)

	// Call heapify_bottom_up to bubble up the last element.
	heapify_bottom_up(&h)

	// Expected array after bubbling up: [0, 1, 10, 20, 5]
	expected := []int{0, 1, 10, 20, 5}

	if len(h.items) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(h.items))
	}
	for i, v := range expected {
		if h.items[i] != v {
			t.Errorf("at index %d expected %d got %d", i, v, h.items[i])
		}
	}
}

func TestHeapifyBottomUpNoBubble(t *testing.T) {
	// Create a valid min-heap: [0, 5, 10, 20]
	// Then simulate inserting a new element "30" that does not need to bubble up.
	h := Heap[int]{
		custom_is_lower_fn: is_lower,
		items:             []int{0, 5, 10, 20},
	}
	h.items = append(h.items, 30)

	// Call heapify_bottom_up.
	heapify_bottom_up(&h)

	// Expected array remains unchanged.
	expected := []int{0, 5, 10, 20, 30}

	if len(h.items) != len(expected) {
		t.Fatalf("expected length %d, got %d", len(expected), len(h.items))
	}
	for i, v := range expected {
		if h.items[i] != v {
			t.Errorf("at index %d expected %d got %d", i, v, h.items[i])
		}
	}
}

