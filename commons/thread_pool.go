package commons

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type WriteOnlyThreadPool[T any] struct {
	inputChannel          chan T
	workerFn              func(chan T, *sync.WaitGroup)
	workloadFactorSamples []float64
	waiting               sync.WaitGroup
	maxWorkers            int
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
	threadPool.workerFn = setupWorkerFunction(workerFn)
	threadPool.closed = false

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
		return nil
	case <-time.After(time.Second * 3):
		return fmt.Errorf("submit timeout - workers may be overloaded")
	}
}

func (tp *WriteOnlyThreadPool[T]) SyncAndClose() {
	tp.closed = true

	for len(tp.inputChannel) != 0 {
		time.Sleep(10 * time.Millisecond)
	}

	close(tp.inputChannel)

	tp.waiting.Wait()
}

func (tp *WriteOnlyThreadPool[T]) Sync() {
	for len(tp.inputChannel) != 0 {
		time.Sleep(10 * time.Millisecond)
	}
}

func setupWorkerFunction[T any](fn func(T)) func(chan T, *sync.WaitGroup) {
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

func (tp *WriteOnlyThreadPool[T]) addNewWorker() {
	tp.waiting.Add(1)
	go tp.workerFn(tp.inputChannel, &tp.waiting)
}
