package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"sync"
)

type FileHeap struct {
	heap          ds.Heap[commons.File]
	hash_registry *ds.Flyweight[string]
	size_filter   sync.Map
}

func new_file_heap(compare_fn func(commons.File, commons.File) bool, registry *ds.Flyweight[string]) *FileHeap {
	file_heap := FileHeap{}

	file_heap.heap = ds.Heap[commons.File]{}
	file_heap.hash_registry = registry

	file_heap.heap.Compare_fn(compare_fn)

	return &file_heap
}

func refine_and_push_file_into_heap(file commons.File, file_chan chan<- commons.File, flyweight *ds.Flyweight[string]) {
	if file.Hash.Value() == "" {
		hash, err := commons.CalculateHash(file.Name)
		if err != nil {
			panic(err)
		}

		file.Hash = flyweight.Instance(hash)
	}

	file_chan <- file
}

func get_file_hash_thread_fn(file_chan chan<- commons.File, flyweight *ds.Flyweight[string]) func(commons.File) {
	return func(obj commons.File) {
		refine_and_push_file_into_heap(obj, file_chan, flyweight)
	}
}

func file_channel_consumer(channel chan commons.File, waitgroup *sync.WaitGroup, heap *ds.Heap[commons.File]) {
	defer waitgroup.Done()
	for data := range channel {
		heap.Push(data)
	}
}

func (file_heap *FileHeap) filter_heap(filter_fn func(commons.File, commons.File) bool, registry *ds.Flyweight[string]) *FileHeap {
	var current commons.File
	var last commons.File

	output := new_file_heap(commons.HashDescending, registry)
	output.hash_registry = file_heap.hash_registry

	total := float64(file_heap.heap.Size())
	processed := 0.0

	duplicate_flag := false

	file_channel := make(chan commons.File)
	target_fn := get_file_hash_thread_fn(file_channel, output.hash_registry)
	file_threadpool, err := commons.NewWorkerPool(target_fn)

	if err != nil {
		panic(err)
	}

	output_waitgroup := sync.WaitGroup{}
	output_waitgroup.Add(1)
	go file_channel_consumer(file_channel, &output_waitgroup, &output.heap)

	ui.Register_line("cleanup-stage", "Removing unique entries %s ... %.1f %%")

	if !file_heap.heap.Empty() {
		current = file_heap.heap.Pop()
		processed += 1.0
	}

	for !file_heap.heap.Empty() {
		last = current
		current = file_heap.heap.Pop()
		processed += 1.0

		if filter_fn(current, last) {
			duplicate_flag = true
			file_threadpool.Submit(last)
		} else if duplicate_flag {
			duplicate_flag = false
			file_threadpool.Submit(last)
		} else {
			duplicate_flag = false
		}

		ui.Update_line("cleanup-stage", "cleanup-stage", (processed/total)*100)
	}

	if duplicate_flag {
		file_threadpool.Submit(current)
	}

	file_threadpool.SyncAndClose()
	close(file_channel)
	output_waitgroup.Wait()

	return output
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
			ui.Print_not_registered("file: %s", last_file_seen)
		} else {
			if is_duplicate {
				ui.Print_not_registered("file: %s", last_file_seen)
			}
		}

		is_duplicate = files_are_equal

	}

	if is_duplicate {
		ui.Print_not_registered("file: %s", last_file_seen)
	}
}
