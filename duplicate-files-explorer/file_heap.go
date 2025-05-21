package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
)

type FileHeap struct {
	heap           *ds.Heap[commons.File]
	pending_insert *ds.AtomicCounter
}

func build_new_file_heap() *FileHeap {
	file_heap := FileHeap{}
	new_heap := ds.Heap[commons.File]{}

	file_heap.heap = &new_heap
	file_heap.heap.Compare_fn(commons.Lower)
	file_heap.pending_insert = ds.Build_new_atomic_counter()

	if file_heap.heap == nil || file_heap.pending_insert == nil {
		panic("error wile creating new file heap object")
	}

	return &file_heap
}

func refine_and_push_file_into_heap(file commons.File, file_heap *FileHeap) {
	file_heap.pending_insert.Increment()

	hash, err := commons.Hash(file.Name, file.Size)

	if err == nil {
		file.Hash = hash
		file_heap.heap.Push(file)
	}

	file_heap.pending_insert.Decrement()
}

func get_file_hash_thread_fn(file_heap *FileHeap) func(commons.File) {
	return func(obj commons.File) {
		refine_and_push_file_into_heap(obj, file_heap)
	}
}

func build_duplicate_entries_heap(file_heap *ds.Heap[commons.File]) *ds.Heap[commons.File] {
	var last_file_seen commons.File
	var current_file commons.File
	var files_are_equal bool
	var file_thread_pool commons.WriteOnlyThreadPool[commons.File] = commons.WriteOnlyThreadPool[commons.File]{}

	refined_file_heap := build_new_file_heap()
	last_seen_was_a_duplicate := false

	total_entries := float64(file_heap.Size())
	ignored_files_counter := 0
	processed_entries := float64(0)

	file_thread_pool.Init(get_file_hash_thread_fn(refined_file_heap))

	main_ui.Register_line("cleanup-stage", "Removing unique entries %s ... %.1f %%")

	if !file_heap.Empty() {
		current_file = file_heap.Pop()
	}

	for !file_heap.Empty() {
		last_file_seen = current_file
		current_file = file_heap.Pop()
		processed_entries++
		files_are_equal = commons.Equal(current_file, last_file_seen)

		if files_are_equal {
			last_seen_was_a_duplicate = true
			file_thread_pool.Submit(last_file_seen)
		} else {
			if last_seen_was_a_duplicate {
				file_thread_pool.Submit(last_file_seen)
			}
			last_seen_was_a_duplicate = false
			ignored_files_counter++
		}

		main_ui.Update_line("cleanup-stage", "cleanup-stage", (processed_entries/total_entries)*100)
	}

	file_thread_pool.Sync_and_close()

	return refined_file_heap.heap
}

func display_duplicate_file_info(file_heap *ds.Heap[commons.File]) {
	var last_file_seen commons.File
	var current_file commons.File
	var files_are_equal bool

	is_duplicate := false

	if !file_heap.Empty() {
		current_file = file_heap.Pop()
	}

	for !file_heap.Empty() {
		last_file_seen = current_file
		current_file = file_heap.Pop()
		files_are_equal = commons.Equal(current_file, last_file_seen)

		if files_are_equal {
			commons.Print_not_registered(main_ui, "file: %s\n", last_file_seen)
			is_duplicate = true
		} else {
			if is_duplicate {
				commons.Print_not_registered(main_ui, "file: %s\n", last_file_seen)
			}
			is_duplicate = false
		}

	}

	if is_duplicate {
		commons.Print_not_registered(main_ui, "file: %s\n", last_file_seen)
	}
}
