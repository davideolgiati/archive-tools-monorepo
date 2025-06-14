package fuzztests

import (
	"archive-tools-monorepo/dataStructures"
	"testing"
)


func FuzzQueue(f *testing.F) {
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

		queue := dataStructures.Queue[*string]{}
		queue.Init()

		offset := 0
		treshold := 29

		for i := 0; i < target; i++ {
			array[i] = getRandomString()
			queue.Push(array[i])
			if queue.Size() == treshold {
				for !queue.Empty() {
					data, err := queue.Pop()

					if err != nil {
						t.Errorf("Error while popping from queue: %v", err)
					}

					if *data != *array[offset] {
						t.Errorf("Error while popping from queue:\n\texpected: %v\n\t got: %v", *array[i], *data)
					}

					if data != array[offset] {
						t.Errorf("Error while popping from queue:\n\texpected: %v\n\t got: %v", array[i], data)
					}

					offset++
				}
				treshold = (treshold * 2) + 1
			}
			if queue.Size() != ((i + 1) - offset) {
				t.Errorf("Error in queue size:\n\texpected: %v\n\t got: %v", ((i + 1) - offset), queue.Size())
			}
		}

		for i := offset; i < target; i++ {
			data, err := queue.Pop()

			if err != nil {
				t.Errorf("Error while popping from queue: %v", err)
			}

			if data != array[i] {
				t.Errorf("Error while popping from queue:\n\texpected: %v\n\t got: %v", array[i], data)
			}

			if *data != *array[i] {
				t.Errorf("Error while popping from queue:\n\texpected: %v\n\t got: %v", *array[i], *data)
			}
		}

		if !queue.Empty() {
			t.Errorf("Error! Queue not empty!")
		}

		if queue.Size() != 0 {
			t.Errorf("Error! Queue empty but size != 0!")
		}
	})
}
