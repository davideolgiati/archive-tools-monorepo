package commons

import "sync"

type Flyweight[T comparable] struct {
	cache map[T]*T
	mutex sync.Mutex
}

func (fw *Flyweight[T]) Init() {
	fw.cache = make(map[T]*T)
}

func (fw *Flyweight[T]) Cache_reference(data T) *T {
	fw.mutex.Lock()
	defer fw.mutex.Unlock()
	
	if _, ok := fw.cache[data]; !ok {
		value := data
		fw.cache[data] = &value
	}

	return fw.cache[data]
}