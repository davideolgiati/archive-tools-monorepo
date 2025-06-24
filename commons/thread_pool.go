package commons

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

const defaultSampleSize = 10

type poolConfiguration[T any] struct {
	workerFunction func(*poolSharedResources[T])
	maxWorkers     int
}

type poolStatus struct {
	poolLoad      []float64
	activeThreads atomic.Int64
	isClosed      bool
}

type poolSharedResources[T any] struct {
	inputChannel chan T
	waitingGroup sync.WaitGroup
}

type WriteOnlyThreadPool[T any] struct {
	configuration poolConfiguration[T]
	status        poolStatus
	shared        poolSharedResources[T]
}

func NewWorkerPool[T any](workerFn func(T)) (*WriteOnlyThreadPool[T], error) {
	threadPool := &WriteOnlyThreadPool[T]{}
	workersCount := runtime.NumCPU()
	inputChannel := make(chan T)
	sampleArray := make([]float64, defaultSampleSize)

	for i := range defaultSampleSize {
		sampleArray[i] = 0.5
	}

	if workerFn == nil {
		return threadPool, errors.New("target function for thread pool can't be null")
	}

	if workersCount < 1 {
		return threadPool, errors.New("error while looking for CPU info in threadpool setup")
	}

	threadPool.status = poolStatus{}
	threadPool.status.activeThreads.Store(0)
	threadPool.status.isClosed = false
	threadPool.status.poolLoad = sampleArray

	threadPool.configuration = poolConfiguration[T]{}
	threadPool.configuration.maxWorkers = workersCount
	threadPool.configuration.workerFunction = setupWorkerFunction(workerFn, &threadPool.status.activeThreads)

	threadPool.shared.inputChannel = inputChannel

	for range threadPool.configuration.maxWorkers {
		threadPool.addNewWorker()
	}

	return threadPool, nil
}

func (tp *WriteOnlyThreadPool[T]) Submit(data T) error {
	if tp.status.isClosed {
		return errors.New("send on closed thread pool")
	}

	select {
	case tp.shared.inputChannel <- data:
		tp.status.activeThreads.Add(1)
		return nil
	case <-time.After(time.Second * 3):
		return errors.New("submit timeout - workers may be overloaded")
	}
}

func (tp *WriteOnlyThreadPool[T]) Release() {
	tp.status.isClosed = true

	tp.Wait()
	close(tp.shared.inputChannel)

	tp.shared.waitingGroup.Wait()
}

func (tp *WriteOnlyThreadPool[T]) Wait() {
	for tp.status.activeThreads.Load() > 0 {
		time.Sleep(1 * time.Millisecond)
	}
}

func setupWorkerFunction[T any](fn func(T), activeTasks *atomic.Int64) func(*poolSharedResources[T]) {
	return func(shared *poolSharedResources[T]) {
		if shared == nil {
			panic("poolSharedResources pointer is nil")
		}

		defer shared.waitingGroup.Done()

		for obj := range shared.inputChannel {
			fn(obj)
			activeTasks.Add(-1)
		}
	}
}

func (tp *WriteOnlyThreadPool[T]) addNewWorker() {
	tp.shared.waitingGroup.Add(1)
	go tp.configuration.workerFunction(&tp.shared)
}
