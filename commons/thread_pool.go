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
	shared        poolSharedResources[T]
	status        poolStatus
}

func NewWorkerPool[T any](workerFn func(T) error) (*WriteOnlyThreadPool[T], error) {
	workersCount := runtime.NumCPU()
	inputChannel := make(chan T)
	sampleArray := make([]float64, defaultSampleSize)

	for i := range defaultSampleSize {
		sampleArray[i] = 0.5
	}

	if workerFn == nil {
		return nil, errors.New("target function for thread pool can't be null")
	}

	if workersCount < 1 {
		return nil, errors.New("error while looking for CPU info in threadpool setup")
	}

	threadPool := &WriteOnlyThreadPool[T]{
		status: poolStatus{
			activeThreads: atomic.Int64{},
			isClosed:      false,
			poolLoad:      sampleArray,
		},
		configuration: poolConfiguration[T]{
			maxWorkers:     workersCount,
			workerFunction: nil,
		},
		shared: poolSharedResources[T]{
			inputChannel: inputChannel,
			waitingGroup: sync.WaitGroup{},
		},
	}

	threadPool.status.activeThreads.Store(0)
	threadPool.configuration.workerFunction = setupWorkerFunction(workerFn, &threadPool.status.activeThreads)

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

func setupWorkerFunction[T any](fn func(T) error, activeTasks *atomic.Int64) func(*poolSharedResources[T]) {
	return func(shared *poolSharedResources[T]) {
		var err error
		if shared == nil {
			return // TODO: this need to be fixed
		}

		defer shared.waitingGroup.Done()

		for obj := range shared.inputChannel {
			err = fn(obj)
			activeTasks.Add(-1)

			if err != nil {
				continue // TODO: this need to be fixed
			}
		}
	}
}

func (tp *WriteOnlyThreadPool[T]) addNewWorker() {
	tp.shared.waitingGroup.Add(1)
	go tp.configuration.workerFunction(&tp.shared) // TODO: return error on an erro channel
}
