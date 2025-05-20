package commons

import (
	"runtime"
	"sync"
	"time"
)

type WriteOnlyThreadPool[T any] struct {
	input_channel   chan T
	waiting         sync.WaitGroup
	max_workers     int
	min_workers     int
	current_workers int
	worker_fn       func(chan T, chan bool, *sync.WaitGroup)
	stop_channels 	[]chan bool
}

func (tp *WriteOnlyThreadPool[T]) Init(fn func(T)) {
	tp.max_workers = runtime.NumCPU() * 4
	tp.min_workers = 1
	tp.current_workers = 0
	
	tp.input_channel = make(chan T, tp.max_workers * 2)
	tp.stop_channels = make([]chan bool, tp.max_workers)
	for i := 0; i < tp.max_workers; i++ {
		tp.stop_channels[i] = make(chan bool)
	}

	tp.worker_fn = setup_fn(fn)

	go tp.manage_thread_pool()
}

func (tp *WriteOnlyThreadPool[T]) Submit(data T) {
	tp.input_channel <- data
}

func (tp *WriteOnlyThreadPool[T]) Sync_and_close() {
	close(tp.input_channel)
	tp.waiting.Wait()
}

func setup_fn[T any](fn func(T)) func(chan T, chan bool, *sync.WaitGroup) {
	return func(in chan T, stop chan bool, wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			select {
			case obj, ok := <-in:
				if !ok {
					return
				}

				fn(obj)
			case <-stop:
				return
			}
		}
	}
}

func (tp *WriteOnlyThreadPool[T]) add_new_worker() {
	tp.current_workers += 1
	tp.waiting.Add(1)
	go tp.worker_fn(tp.input_channel, tp.stop_channels[tp.current_workers - 1], &tp.waiting)
}

func (tp *WriteOnlyThreadPool[T]) stop_last_worker() {
	tp.stop_channels[tp.current_workers - 1] <- true
	tp.current_workers = tp.current_workers - 1
}

func (tp * WriteOnlyThreadPool[T]) manage_thread_pool() {
	for {
		workload_factor := float64(len(tp.input_channel)) / float64(tp.max_workers)
		if workload_factor > 0.7 && tp.current_workers < tp.max_workers {
			tp.add_new_worker()
		} else if workload_factor < 0.3 && tp.current_workers > tp.min_workers {
			tp.stop_last_worker()
		}

		time.Sleep(300 * time.Millisecond)
	}
}
