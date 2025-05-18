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
	defer counter.mutex.Unlock()

	counter.value += 1
}

func (counter *AtomicCounter) Decrement() {
	counter.mutex.Lock()
	defer counter.mutex.Unlock()

	counter.value -= 1
}

func (counter *AtomicCounter) Value() int64 {
	counter.mutex.Lock()
	defer counter.mutex.Unlock()
	
	output := counter.value

	return output
}
