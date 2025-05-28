package commons

import (
	"runtime"
	"sync"
	"time"
)

type WriteOnlyThreadPool[T any] struct {
	input_channel           chan T
	worker_fn               func(chan T, chan bool, *sync.WaitGroup)
	stop_channels           []chan bool
	workload_factor_samples []float64
	waiting                 sync.WaitGroup
	max_workers             int
	min_workers             int
	current_workers         int
	mutex                   sync.Mutex
}

func (tp *WriteOnlyThreadPool[T]) Init(fn func(T)) {
	tp.max_workers = runtime.NumCPU()
	tp.min_workers = 1
	tp.current_workers = 0

	tp.input_channel = make(chan T, tp.max_workers)
	tp.stop_channels = make([]chan bool, tp.max_workers)
	for i := 0; i < tp.max_workers; i++ {
		tp.stop_channels[i] = make(chan bool)
	}

	tp.workload_factor_samples = make([]float64, 10)

	for i := 0; i < 10; i++ {
		tp.workload_factor_samples[i] = 0.5
	}

	tp.worker_fn = setup_fn(fn)

	for i := 0; i < (tp.max_workers/2) + 1; i++ {
		tp.add_new_worker()
	}

	go tp.sample_pool_usage()

	go tp.manage_thread_pool()
}

func (tp *WriteOnlyThreadPool[T]) Submit(data T) {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	tp.input_channel <- data
}

func (tp *WriteOnlyThreadPool[T]) Sync_and_close() {
	for len(tp.input_channel) > 0 {
		time.Sleep(1 * time.Second)
	}
	
	close(tp.input_channel)
	
	tp.waiting.Wait()
}

func (tp *WriteOnlyThreadPool[T]) sample_pool_usage() {
	var index int = 0
	for {
		index = (index + 1) % 10
		tp.mutex.Lock()
		tp.workload_factor_samples[index] = float64(len(tp.input_channel)) / float64(tp.max_workers)
		tp.mutex.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
}

func setup_fn[T any](fn func(T)) func(chan T, chan bool, *sync.WaitGroup) {
	return func(in chan T, stop chan bool, wg *sync.WaitGroup) {
		if wg == nil {
			panic("waitgroup pointer is nil")
		}

		for {
			select {
			case obj, ok := <-in:
				if !ok {
					wg.Done()
					return
				}

				fn(obj)
			case <-stop:
				wg.Done()
				return
			}
		}
	}
}

func (tp *WriteOnlyThreadPool[T]) add_new_worker() {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()

	tp.current_workers += 1
	tp.waiting.Add(1)
	go tp.worker_fn(tp.input_channel, tp.stop_channels[tp.current_workers-1], &tp.waiting)
}

func (tp *WriteOnlyThreadPool[T]) stop_last_worker() {
	tp.mutex.Lock()
	defer tp.mutex.Unlock()
	
	tp.stop_channels[tp.current_workers-1] <- true
	tp.current_workers = tp.current_workers - 1
}

func (tp *WriteOnlyThreadPool[T]) manage_thread_pool() {
	for tp.current_workers > 0 {
		workload_factor := mean(tp.workload_factor_samples, &tp.mutex)
		if workload_factor > 0.7 && tp.current_workers < tp.max_workers {
			tp.add_new_worker()
		} else if workload_factor < 0.3 && tp.current_workers > 0 {
			tp.stop_last_worker()
		}

		time.Sleep(300 * time.Millisecond)
	}
}

func mean(data []float64, mutex *sync.Mutex) float64 {
	mutex.Lock()
	defer mutex.Unlock()
	
	if len(data) == 0 {
		return 0
	}
	var sum float64
	for _, d := range data {
		sum += d
	}
	return sum / float64(len(data))
}
