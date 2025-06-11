package commons

import (
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type WriteOnlyThreadPool[T any] struct {
	inputChannel          chan T
	workerFn              func(chan T, *sync.WaitGroup)
	workloadFactorSamples []float64
	waitingGroup          sync.WaitGroup
	maxWorkers            int
	activeTasks           atomic.Int64
	closed                bool
}

func NewWorkerPool[T any](workerFn func(T)) (*WriteOnlyThreadPool[T], error) {
	threadPool := &WriteOnlyThreadPool[T]{}
	workersCount := runtime.NumCPU()
	inputChannel := make(chan T)
	sampleArray := make([]float64, 10)

	for i := 0; i < 10; i++ {
		sampleArray[i] = 0.5
	}

	if workerFn == nil {
		return threadPool, fmt.Errorf("target function for thread pool can't be null")
	}

	if workersCount < 1 {
		return threadPool, fmt.Errorf("error while looking for CPU info in threadpool setup")
	}

	threadPool.maxWorkers = workersCount
	threadPool.inputChannel = inputChannel
	threadPool.workloadFactorSamples = sampleArray
	threadPool.workerFn = setupWorkerFunction(workerFn, &threadPool.activeTasks)
	threadPool.closed = false

	threadPool.activeTasks = atomic.Int64{}
	threadPool.activeTasks.Store(0)

	for i := 0; i < threadPool.maxWorkers; i++ {
		threadPool.addNewWorker()
	}

	return threadPool, nil
}

func (tp *WriteOnlyThreadPool[T]) Submit(data T) error {
	if tp.closed {
		return fmt.Errorf("send on closed thread pool")
	}

	select {
	case tp.inputChannel <- data:
		tp.activeTasks.Add(1)
		return nil
	case <-time.After(time.Second * 3):
		return fmt.Errorf("submit timeout - workers may be overloaded")
	}
}

func (tp *WriteOnlyThreadPool[T]) Release() {
	tp.closed = true

	tp.Wait()
	close(tp.inputChannel)

	tp.waitingGroup.Wait()
}

func (tp *WriteOnlyThreadPool[T]) Wait() {
	for tp.activeTasks.Load() > 0 {
		time.Sleep(1 * time.Millisecond)
	}
}

func setupWorkerFunction[T any](fn func(T), activeTasks *atomic.Int64) func(chan T, *sync.WaitGroup) {
	return func(in chan T, wg *sync.WaitGroup) {
		if wg == nil {
			panic("waitgroup pointer is nil")
		}

		defer wg.Done()

		for obj := range in {
			fn(obj)
			activeTasks.Add(-1)
		}
	}
}

func (tp *WriteOnlyThreadPool[T]) addNewWorker() {
	tp.waitingGroup.Add(1)
	go tp.workerFn(tp.inputChannel, &tp.waitingGroup)
}
