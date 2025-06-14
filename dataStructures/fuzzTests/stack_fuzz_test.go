package fuzztests

import (
	"archive-tools-monorepo/dataStructures"
	"testing"
)

func FuzzStack(f *testing.F) {
	testcases := []int{
		10, 100, 1000,
	}
	for _, tc := range testcases {
		f.Add(tc)
	}
	f.Fuzz(func(t *testing.T, target int) {
		if target < 0 {
			target *= -1
		}
		array := make([]*string, target)

		stack := dataStructures.Stack[*string]{}

		offset := 0

		for i := 0; i < target; i++ {
			array[i] = getRandomString()
			stack.Push(array[i])
			if stack.Size() != ((i + 1) - offset) {
				t.Errorf("Error in queue size:\n\texpected: %v\n\t got: %v", ((i + 1) - offset), stack.Size())
			}
		}

		for i := target-1; i >= 0; i-- {
			data := stack.Pop()

			if data != array[i] {
				t.Errorf("Error while popping from queue:\n\texpected: %v\n\t got: %v", array[i], data)
			}

			if *data != *array[i] {
				t.Errorf("Error while popping from queue:\n\texpected: %v\n\t got: %v", *array[i], *data)
			}
		}

		if !stack.Empty() {
			t.Errorf("Error! Queue not empty!")
		}

		if stack.Size() != 0 {
			t.Errorf("Error! Queue empty but size != 0!")
		}
	})
}
