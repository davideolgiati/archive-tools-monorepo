package datastructures

type IDataStructure[T any] interface {
	Empty() bool
	Size() int
	Push(data T) error
	Pop() (T, error)
	Peak() *T
}
