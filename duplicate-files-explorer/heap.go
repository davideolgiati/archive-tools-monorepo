package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"fmt"
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

func refine_and_push_file_into_heap(file *commons.File, file_heap *FileHeap) {
	ds.Increment(&file_heap.pending_insert)

	if file.Size > 4000 {
		hash_channel := make(chan string)
	
		go commons.Hash_file(file.Name, false, hash_channel)
		file.Hash = <-hash_channel
	}

	ds.Push_into_heap(&file_heap.heap, file)

	ds.Decrement(&file_heap.pending_insert)
}

func new_file_heap() *FileHeap {
	new_file_heap := FileHeap{}
	ds.Set_compare_fn(&new_file_heap.heap, commons.Compare_file_hashes)
	new_file_heap.pending_insert = *ds.Create_new_atomic_counter()

	return &new_file_heap
}

func build_duplicate_entries_heap(file_heap *ds.Heap[commons.File]) *ds.Heap[commons.File] {
	var last_file_seen *commons.File
	var current_file *commons.File
	var back_pressure time.Duration
	
	refined_file_heap := new_file_heap()
	is_duplicate := false
	
	input_heap_size := ds.Get_heap_size(file_heap)
	ignored_files_counter := 0
	processed_files_counter := 0

	commons.Register_new_line("heap-line", main_ui)

	if !ds.Is_heap_empty(file_heap) {
		current_file = ds.Pop_from_heap(file_heap)
	}

	for !ds.Is_heap_empty(file_heap) {
		last_file_seen = current_file
		current_file = ds.Pop_from_heap(file_heap)
		processed_files_counter++

		if current_file.Hash == last_file_seen.Hash || is_duplicate {
			is_duplicate = current_file.Hash == last_file_seen.Hash
			go refine_and_push_file_into_heap(last_file_seen, refined_file_heap)
		} else {
			ignored_files_counter++
		}

		commons.Print_to_line(
			main_ui, "heap-line",
			"Removing unique entries stage-1 ... %.1f %%",
			(float64(processed_files_counter)/float64(input_heap_size))*100,
		)

		back_pressure = compute_back_pressure(&refined_file_heap.pending_insert)
		time.Sleep(back_pressure)
	}

	fmt.Print("\n")

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
