package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"sync"
)

type FileHeap struct {
	heap ds.Heap[commons.File]
	hash_registry commons.Flyweight[string]
}

func build_new_file_heap(compare_fn func(commons.File, commons.File) bool) *FileHeap {
	file_heap := FileHeap{}

	file_heap.heap = ds.Heap[commons.File]{}
	file_heap.hash_registry = commons.Flyweight[string]{}

	file_heap.hash_registry.Init()

	file_heap.heap.Compare_fn(compare_fn)

	return &file_heap
}

func refine_and_push_file_into_heap(file commons.File, file_chan chan<- commons.File, flyweight *commons.Flyweight[string]) {
	hash, err := commons.Hash(file.Name, file.Size)

	if err == nil {
		file.Hash = flyweight.Cache_reference(hash)
		file_chan <- file
	} else {
		panic(err)
	}
}

func get_file_hash_thread_fn(file_chan chan<- commons.File, flyweight *commons.Flyweight[string]) func(commons.File) {
	return func(obj commons.File) {
		refine_and_push_file_into_heap(obj, file_chan, flyweight)
	}
}

func build_duplicate_entries_heap(file_heap *FileHeap) *FileHeap {
	var last_file_seen commons.File
	var current_file commons.File
	
	var file_thread_pool commons.WriteOnlyThreadPool[commons.File] = commons.WriteOnlyThreadPool[commons.File]{}

	output_channel := make(chan commons.File)
	output_wg := sync.WaitGroup{}

	refined_file_heap := build_new_file_heap(commons.HashDescending)
	last_seen_was_a_duplicate := false

	total_entries := float64(file_heap.heap.Size())
	processed_entries := float64(0)

	file_thread_pool.Init(get_file_hash_thread_fn(output_channel, &refined_file_heap.hash_registry))

	main_ui.Register_line("cleanup-stage", "Removing unique entries %s ... %.1f %%")

	output_wg.Add(1)

	go func() {
		defer output_wg.Done()
		for data := range output_channel {
			refined_file_heap.heap.Push(data)
		}
	}()

	if !file_heap.heap.Empty() {
		current_file = file_heap.heap.Pop()
	}

	for !file_heap.heap.Empty() {
		last_file_seen = current_file
		current_file = file_heap.heap.Pop()
		processed_entries++

		if commons.EqualBySize(current_file, last_file_seen) {
			last_seen_was_a_duplicate = true
			file_thread_pool.Submit(last_file_seen)
		} else {
			if last_seen_was_a_duplicate {
				file_thread_pool.Submit(last_file_seen)
			}
			last_seen_was_a_duplicate = false
		}

		main_ui.Update_line("cleanup-stage", "cleanup-stage", (processed_entries/total_entries)*100)
	}

	if last_seen_was_a_duplicate {
		file_thread_pool.Submit(last_file_seen)
	}

	file_thread_pool.Sync_and_close()

	close(output_channel)

	output_wg.Wait()

	return refined_file_heap
}

func display_duplicate_file_info(file_heap *FileHeap) {
	var last_file_seen commons.File
	var current_file commons.File
	var files_are_equal bool

	is_duplicate := false

	if !file_heap.heap.Empty() {
		current_file = file_heap.heap.Pop()
	}

	for !file_heap.heap.Empty() {
		last_file_seen = current_file
		current_file = file_heap.heap.Pop()
		files_are_equal = commons.EqualByHash(current_file, last_file_seen)

		if files_are_equal {
			commons.Print_not_registered(main_ui, "file: %s", last_file_seen)
		} else {
			if is_duplicate {
				commons.Print_not_registered(main_ui, "file: %s", last_file_seen)
			}
		}

		is_duplicate = files_are_equal

	}

	if is_duplicate {
		commons.Print_not_registered(main_ui, "file: %s", last_file_seen)
	}
}
