package ds

import (
	"sync"
)

type AtomicCounter struct {
	mutex sync.Mutex
	value int64
}

func Build_new_atomic_counter() *AtomicCounter {
	output := AtomicCounter{}
	output.value = 0

	return &output
}

func (counter *AtomicCounter) Increment() {
	counter.mutex.Lock()
	counter.value += 1
	counter.mutex.Unlock()
}

func (counter *AtomicCounter) Decrement() {
	counter.mutex.Lock()
	counter.value -= 1
	counter.mutex.Unlock()
}

func (counter *AtomicCounter) Value() int64 {
	counter.mutex.Lock()
	output := counter.value
	counter.mutex.Unlock()

	return output
}
