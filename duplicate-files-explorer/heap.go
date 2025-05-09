package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"time"
)

func compute_back_pressure(queue *ds.AtomicCounter) time.Duration {
	queue_size := ds.Get_counter_value(queue)

	if queue_size < 500 {
		return 0 * time.Millisecond
	}

	if queue_size < 1000 {
		return 1 * time.Millisecond
	}

	if queue_size < 2000 {
		return 2 * time.Millisecond
	}

	return 3 * time.Millisecond
}

func refine_and_push_file_into_heap(file *commons.File, file_heap *FileHeap, lazy bool) {
	ds.Increment(&file_heap.pending_insert)

	file.Hash = commons.Hash_file(file.Name, lazy)
	ds.Push_into_heap(&file_heap.heap, file)

	ds.Decrement(&file_heap.pending_insert)
}

func build_new_file_heap() *FileHeap {
	file_heap := FileHeap{}

	ds.Set_compare_fn(&file_heap.heap, commons.Compare_files)
	file_heap.pending_insert = *ds.Build_new_atomic_counter()

	return &file_heap
}

func build_duplicate_entries_heap(file_heap *ds.Heap[commons.File], lazy bool) *ds.Heap[commons.File] {
	var last_file_seen *commons.File
	var current_file *commons.File
	var back_pressure time.Duration
	var line_id string

	refined_file_heap := build_new_file_heap()
	is_duplicate := false

	input_heap_size := ds.Get_heap_size(file_heap)
	ignored_files_counter := 0
	processed_files_counter := 0

	if lazy {
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

		if (current_file.Hash == last_file_seen.Hash && current_file.Size == last_file_seen.Size) || is_duplicate {
			is_duplicate = (current_file.Hash == last_file_seen.Hash && current_file.Size == last_file_seen.Size)
			go refine_and_push_file_into_heap(last_file_seen, refined_file_heap, lazy)
		} else {
			ignored_files_counter++
		}

		commons.Print_to_line(
			main_ui, line_id,
			"Removing unique entries %s ... %.1f %%", line_id,
			(float64(processed_files_counter)/float64(input_heap_size))*100,
		)

		back_pressure = compute_back_pressure(&refined_file_heap.pending_insert)
		time.Sleep(back_pressure)
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
			print_file_details_to_stdout(last_file_seen)
			is_duplicate = current_file.Hash == last_file_seen.Hash
		}

	}
}
