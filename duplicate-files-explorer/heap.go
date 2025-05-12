package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"fmt"
	"runtime"
	"time"
)

func apply_back_pressure(queue *ds.AtomicCounter) {
	queue_size := ds.Get_counter_value(queue)

	if queue_size > int64(runtime.NumCPU()) {
		time.Sleep(100 * time.Microsecond)
	}
}

func refine_and_push_file_into_heap(file *commons.File, file_heap *FileHeap, lazy bool) {
	ds.Increment(&file_heap.pending_insert)

	hash, err := commons.Hash(&file.Name, file.Size, lazy)

	if err == nil {
		file.Hash = hash
		ds.Push_into_heap(&file_heap.heap, file)
	}

	ds.Decrement(&file_heap.pending_insert)
}

func build_duplicate_entries_heap(file_heap *ds.Heap[commons.File], lazy_hashing bool) *ds.Heap[commons.File] {
	var last_file_seen *commons.File
	var current_file *commons.File
	var line_id string
	var files_are_equal bool

	refined_file_heap := build_new_file_heap()
	last_seen_was_a_duplicate := false

	input_heap_size := ds.Get_heap_size(file_heap)
	ignored_files_counter := 0
	processed_files_counter := 0

	if lazy_hashing {
		line_id = "stage-1"
	} else {
		line_id = "stage-2"
	}

	commons.Register_new_line(line_id, main_ui)

	if !ds.Is_heap_empty(file_heap) {
		current_file = ds.Pop_from_heap(file_heap)
	}

	for !ds.Is_heap_empty(file_heap) {
		last_file_seen = current_file
		current_file = ds.Pop_from_heap(file_heap)
		processed_files_counter++
		files_are_equal = commons.Equal(current_file, last_file_seen)

		if files_are_equal || last_seen_was_a_duplicate {
			last_seen_was_a_duplicate = files_are_equal
			go refine_and_push_file_into_heap(last_file_seen, refined_file_heap, lazy_hashing)
		} else {
			ignored_files_counter++
		}

		commons.Print_to_line(
			main_ui, line_id,
			"Removing unique entries %s ... %.1f %%", line_id,
			(float64(processed_files_counter)/float64(input_heap_size))*100,
		)

		apply_back_pressure(&refined_file_heap.pending_insert)
	}

	return &refined_file_heap.heap
}

func display_duplicate_file_info(file_heap *ds.Heap[commons.File]) {
	var last_file_seen *commons.File
	var current_file *commons.File

	is_duplicate := false

	if !ds.Is_heap_empty(file_heap) {
		current_file = ds.Pop_from_heap(file_heap)
	}

	for !ds.Is_heap_empty(file_heap) {
		last_file_seen = current_file
		current_file = ds.Pop_from_heap(file_heap)

		if current_file.Hash == last_file_seen.Hash || is_duplicate {
			fmt.Printf("file: %s\n", last_file_seen)
			is_duplicate = current_file.Hash == last_file_seen.Hash
		}

	}
}
