package ds

type Constant[T any] struct {
	ptr *T
}

func (ro Constant[T]) Value() T {
	return *ro.ptr
}

func (ro Constant[T]) Ptr() *T {
	return ro.ptr
}
