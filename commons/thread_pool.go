package commons

import (
	"runtime"
	"sync"
)

type ThreadPool[T any] struct {
	input_channel chan T
	waiting sync.WaitGroup
}

func (tp *ThreadPool[T]) Init(fn func (chan T, *sync.WaitGroup)) {
	tp.input_channel = make(chan T)

	for w := 1; w <= runtime.NumCPU() * 2; w++ {
		tp.waiting.Add(1)
		go fn(tp.input_channel, &tp.waiting)
	}
}

func (tp *ThreadPool[T]) Get_input_channel() chan T {
	return tp.input_channel
}

func (tp *ThreadPool[T]) Sync_and_close() {
	close(tp.input_channel)
	tp.waiting.Wait()
}
