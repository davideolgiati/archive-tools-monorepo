package ds

import "fmt"

type Constant[T any] struct {
	ptr *T
}

func NewConstant[T any](data *T) (Constant[T], error) {
	if data == nil {
		return Constant[T]{}, fmt.Errorf("data pointer is nil")
	}
	return Constant[T]{ptr: data}, nil
}

func (ro Constant[T]) Value() T {
	return *ro.ptr
}

func (ro Constant[T]) Ptr() *T {
	return ro.ptr
}
