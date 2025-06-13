package ds

import "sync"

type Flyweight[T comparable] struct {
	cache sync.Map
}

func (fw *Flyweight[T]) Instance(data T) (Constant[T], error) {
	cache_reference, ok := fw.get(data)

	if ok {
		return cache_reference, nil
	}

	new_entry, err := fw.set(data)

	return new_entry, err
}

func (fw *Flyweight[T]) get(data T) (Constant[T], bool) {
	entry_pointer, ok := fw.cache.Load(data)

	if !ok {
		default_entry := ""
		entry_pointer = &default_entry
	}

	newConstant, err := NewConstant(entry_pointer.(*T))
	
	if err != nil {
		ok = false
	}

	return newConstant, ok
}

func (fw *Flyweight[T]) set(data T) (Constant[T], error) {
	new_entry := data
	ptr := &new_entry

	actual, loaded := fw.cache.LoadOrStore(data, ptr)

	if !loaded {
		return NewConstant(ptr)
	}

	return NewConstant(actual.(*T))
}
