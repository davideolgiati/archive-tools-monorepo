package ds

import "testing"

func TestParentLeftRight(t *testing.T) {
	node_index := 1

	expected_parent := 0
	expected_left := 3
	expected_right := 4 

	actual_parent := get_parent_node_index(node_index)
	actual_left := get_left_node_index(node_index)
	actual_right := get_right_node_index(node_index)
	
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

	actual_parent := get_parent_node_index(node_index)
	actual_left := get_left_node_index(node_index)
	actual_right := get_right_node_index(node_index)
	
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

func is_lower(a int, b int) bool {
	return a < b
}

func TestHeapifyBase(t *testing.T) {
	input_heap := Heap[int]{}

	input_heap.items = []int{2,3,1}
	input_heap.custom_is_lower_fn = is_lower
	
	expected_array := []int{1,3,2,}

	heapify(&input_heap)

	if len(input_heap.items) != len(expected_array) {
		t.Errorf(
			"len(input_heap.items) = %d, expected %d", 
			len(input_heap.items), 3,
		)
	}
	
	for i := 0; i < 3; i++ {
		if input_heap.items[i] != expected_array[i] {
			t.Errorf(
				"input_heap.items[%d] = %d, expected %d", 
				i, input_heap.items[i], expected_array[i],
			)
		}
	}
}

func TestHeapifyLarge(t *testing.T) {
	input_heap := Heap[int]{}

	input_heap.items = []int{2,3,20,30,21,25,31,35,22,23,26,27,32,33,1}
	input_heap.custom_is_lower_fn = is_lower
	
	expected_array := []int{1,3,2,20,30,21,25,31,35,22,23,26,27,32,33}

	heapify(&input_heap)

	if len(input_heap.items) != len(expected_array) {
		t.Errorf(
			"len(input_heap.items) = %d, expected %d", 
			len(input_heap.items), 3,
		)
	}
	
	for i := 0; i < 3; i++ {
		if input_heap.items[i] != expected_array[i] {
			t.Errorf(
				"input_heap.items[%d] = %d, expected %d", 
				i, input_heap.items[i], expected_array[i],
			)
		}
	}
}