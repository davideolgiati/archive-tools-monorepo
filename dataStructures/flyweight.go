package datastructures

import (
	"errors"
	"sync"
)

type Flyweight[T comparable] struct {
	cache sync.Map
}

func (fw *Flyweight[T]) Instance(data T) (Constant[T], error) {
	cacheReference, ok := fw.get(data)

	if ok {
		return cacheReference, nil
	}

	newEntry, err := fw.set(data)

	return newEntry, err
}

func (fw *Flyweight[T]) get(data T) (Constant[T], bool) {
	entryPointer, ok := fw.cache.Load(data)
	if !ok {
		return Constant[T]{nil}, false
	}

	pointer, ok := entryPointer.(*T)
	if !ok {
		return Constant[T]{nil}, false
	}

	newConstant, err := NewConstant(pointer)
	if err != nil {
		return Constant[T]{nil}, false
	}

	return newConstant, true
}

func (fw *Flyweight[T]) set(data T) (Constant[T], error) {
	newEntry := data
	ptr := &newEntry

	actual, loaded := fw.cache.LoadOrStore(data, ptr)

	if !loaded {
		return NewConstant(ptr)
	}

	pointer, ok := actual.(*T)

	if !ok {
		return Constant[T]{}, errors.New("error while casting poinetr")
	}

	return NewConstant(pointer)
}
