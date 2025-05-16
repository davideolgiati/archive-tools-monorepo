package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"runtime"
	"time"
)

type FileHeap struct {
	heap           ds.Heap[commons.File]
	pending_insert ds.AtomicCounter
}

func build_new_file_heap() *FileHeap {
	file_heap := FileHeap{}

	file_heap.heap.Compare_fn(commons.Lower)
	file_heap.pending_insert = *ds.Build_new_atomic_counter()

	return &file_heap
}

func (file_heap *FileHeap) collect() {
	queue_size := file_heap.pending_insert.Value()

	for queue_size > 0 {
		time.Sleep(2 * time.Second)
		queue_size = file_heap.pending_insert.Value()
	}

	if file_heap.pending_insert.Value() != 0 {
		panic("heap.collect() did not wait for all jobs to finish")
	}
}

func refine_and_push_file_into_heap(file commons.File, file_heap *FileHeap, lazy bool) {
	file_heap.pending_insert.Increment()

	hash, err := commons.Hash(&file.Name, file.Size, lazy)

	if err == nil {
		file.Hash = hash
		file_heap.heap.Push(file)
	}

	file_heap.pending_insert.Decrement()
}

func heap_refine_thread_pool(file_heap *FileHeap, lazy_hashing bool, in <-chan commons.File) {
	for obj := range in {
		refine_and_push_file_into_heap(obj, file_heap, lazy_hashing)
	}
}

func build_duplicate_entries_heap(file_heap *ds.Heap[commons.File], lazy_hashing bool) *ds.Heap[commons.File] {
	var last_file_seen commons.File
	var current_file commons.File
	var line_id string
	var files_are_equal bool

	refined_file_heap := build_new_file_heap()
	last_seen_was_a_duplicate := false

	input := make(chan commons.File)

	for w := 1; w <= runtime.NumCPU()*4; w++ {
		go heap_refine_thread_pool(refined_file_heap, lazy_hashing, input)
	}

	total_entries := float64(file_heap.Size())
	ignored_files_counter := 0
	processed_entries := float64(0)

	if lazy_hashing {
		line_id = "stage-1"
	} else {
		line_id = "stage-2"
	}

	main_ui.Register_line(line_id, "Removing unique entries %s ... %.1f %%")

	if !file_heap.Empty() {
		current_file = file_heap.Pop()
	}

	for !file_heap.Empty() {
		last_file_seen = current_file
		current_file = file_heap.Pop()
		processed_entries++
		files_are_equal = commons.Equal(&current_file, &last_file_seen)

		if files_are_equal || last_seen_was_a_duplicate {
			last_seen_was_a_duplicate = files_are_equal
			go refine_and_push_file_into_heap(last_file_seen, refined_file_heap, lazy_hashing)
		} else {
			ignored_files_counter++
		}

		main_ui.Update_line(line_id, line_id, (processed_entries/total_entries)*100)
	}

	refined_file_heap.collect()

	return &refined_file_heap.heap
}

func display_duplicate_file_info(file_heap *ds.Heap[commons.File]) {
	var last_file_seen commons.File
	var current_file commons.File

	is_duplicate := false

	if !file_heap.Empty() {
		current_file = file_heap.Pop()
	}

	for !file_heap.Empty() {
		last_file_seen = current_file
		current_file = file_heap.Pop()

		if current_file.Hash == last_file_seen.Hash || is_duplicate {
			commons.Print_not_registered(main_ui, "file: %s\n", &last_file_seen)
			is_duplicate = current_file.Hash == last_file_seen.Hash
		}

	}

	if is_duplicate {
		commons.Print_not_registered(main_ui, "file: %s\n", &last_file_seen)
	}
}
