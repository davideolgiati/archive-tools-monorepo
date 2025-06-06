package commons

import (
	"runtime"
	"sync"
	"time"
)

type WriteOnlyThreadPool[T any] struct {
	input_channel           chan T
	worker_fn               func(chan T, *sync.WaitGroup)
	workload_factor_samples []float64
	waiting                 sync.WaitGroup
	max_workers             int
	closed                  bool
}

func (tp *WriteOnlyThreadPool[T]) Init(fn func(T)) {
	if fn == nil {
		panic("target function for thread pool can't be null")
	}

	tp.max_workers = runtime.NumCPU()

	if tp.max_workers < 1 {
		panic("error while looking for CPU info in threadpool setup")
	}

	tp.input_channel = make(chan T)

	if tp.input_channel == nil {
		panic("Error while creating input channel for threadpool")
	}

	tp.workload_factor_samples = make([]float64, 10)

	for i := 0; i < 10; i++ {
		tp.workload_factor_samples[i] = 0.5
	}

	tp.worker_fn = setup_fn(fn)

	for i := 0; i < tp.max_workers; i++ {
		tp.add_new_worker()
	}

	tp.closed = false
}

func (tp *WriteOnlyThreadPool[T]) Submit(data T) {
	if tp.closed {
		panic("Send on closed thread pool")
	}

	select {
	case tp.input_channel <- data:
		return
	case <- time.After(time.Second * 3):
		panic("submit timeout - workers may be overloaded")
	}
}

func (tp *WriteOnlyThreadPool[T]) Sync_and_close() {
	tp.closed = true
	
	for len(tp.input_channel) != 0 {
		time.Sleep(10 * time.Millisecond)
	}

	close(tp.input_channel)

	tp.waiting.Wait()
}

func (tp *WriteOnlyThreadPool[T]) Sync() {
	for len(tp.input_channel) != 0 {
		time.Sleep(10 * time.Millisecond)
	}
}

func setup_fn[T any](fn func(T)) func(chan T, *sync.WaitGroup) {
	return func(in chan T, wg *sync.WaitGroup) {
		if wg == nil {
			panic("waitgroup pointer is nil")
		}

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
