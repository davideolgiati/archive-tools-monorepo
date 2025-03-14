package ds

import (
	"sync"
)

type AtomicCounter struct {
	mutex sync.Mutex
	value int64
}

func Create_new_atomic_counter() *AtomicCounter {
	output := AtomicCounter{}
	output.value = 0

	return &output
}

func Increment(counter *AtomicCounter) {
	counter.mutex.Lock()
	counter.value += 1
	counter.mutex.Unlock()
}

func Decrement(counter *AtomicCounter) {
	counter.mutex.Lock()
	counter.value -= 1
	counter.mutex.Unlock()
}

func Get_counter_value(counter *AtomicCounter) int64{
	counter.mutex.Lock()
	output := counter.value
	counter.mutex.Unlock()

	return output
}