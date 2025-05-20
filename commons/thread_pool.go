package commons

import (
	"runtime"
	"sync"
)

type WriteOnlyThreadPool[T any] struct {
	input_channel   chan T
	waiting         sync.WaitGroup
	max_workers     int
	min_workers     int
	current_workers int
	worker_fn       func(chan T, *sync.WaitGroup)
}

func (tp *WriteOnlyThreadPool[T]) Init(fn func(T)) {
	tp.max_workers = runtime.NumCPU() * 4
	tp.min_workers = 1
	tp.current_workers = 0

	tp.worker_fn = setup_fn(fn)

	tp.input_channel = make(chan T)

	for i := 0; i < tp.max_workers; i++ {
		tp.add_new_worker()
	}
}

func setup_fn[T any](fn func(T)) func(chan T, *sync.WaitGroup) {
	return func(in chan T, wg *sync.WaitGroup) {
		defer wg.Done()
		for obj := range in {
			fn(obj)
		}
	}
}

func (tp *WriteOnlyThreadPool[T]) add_new_worker() {
	tp.waiting.Add(1)
	go tp.worker_fn(tp.input_channel, &tp.waiting)
}

func (tp *WriteOnlyThreadPool[T]) Submit(data T) {
	tp.input_channel <- data
}

func (tp *WriteOnlyThreadPool[T]) Sync_and_close() {
	close(tp.input_channel)
	tp.waiting.Wait()
}
