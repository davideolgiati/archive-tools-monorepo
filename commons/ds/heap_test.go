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

