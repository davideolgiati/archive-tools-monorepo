package fuzztests

import (
	"archive-tools-monorepo/dataStructures"
	"fmt"
	"strings"
	"testing"
)

func FuzzQueue(f *testing.F) {
	testCases := []string{
		"p:1;s;o;e",
		"",
		"",
		"o;o;o;o;o;o;o;o;o",
		"p:10",
		"s",
		"e",
		"o",
	}

	for i := 0; i < 35; i++ {
		var data string
		if i%10 == 0 {
			data = "o;"
		} else {
			data = fmt.Sprintf("p:%d;", i)
		}

		testCases[1] += data
	}

	testCases[1] += "e;s"

	testCases[2] = testCases[1] + ";"

	for i := 0; i < 35; i++ {
		testCases[2] += "o;"
	}

	for i := 0; i < 70; i++ {
		var data string
		if i%10 == 0 {
			data = "o;"
		} else {
			data = fmt.Sprintf("p:%d;", i)
		}

		testCases[2] += data
	}

	for _, testCase := range testCases {
		f.Add(testCase)
	}

	f.Fuzz(func(t *testing.T, actions string) {
		var q dataStructures.Queue[string]
		q.Init()

		var model []string

		for i, raw := range strings.Split(actions, ";") {
			switch {
			case strings.HasPrefix(raw, "p:"):
				// Push operation
				valStr := strings.TrimPrefix(raw, "p:")
				q.Push(valStr)
				model = append(model, valStr)
			case raw == "o":
				// Pop operation
				val, err := q.Pop()
				if len(model) == 0 {
					if err == nil {
						t.Fatalf("Step %d: expected error on empty pop, got %v", i, val)
					}
				} else {
					expected := model[0]
					model = model[1:]
					if err != nil {
						t.Fatalf("Step %d: unexpected error: %v", i, err)
					}
					if val != expected {
						t.Fatalf("Step %d: pop mismatch: got %v, expected %v", i, val, expected)
					}
				}
			case raw == "s":
				queueLen := q.Size()
				if queueLen != len(model) {
					t.Fatalf("Step %d: size mismatch: got %v, expected %v", i, queueLen, len(model))
				}
			case raw == "e":
				isEmpty := q.Empty()
				if isEmpty != (len(model) == 0) {
					t.Fatalf("Step %d: size mismatch: got %v, expected %v", i, isEmpty, len(model) == 0)
				}
			default:
				// Unknown op, skip
				continue
			}

			// Invariant checks
			if q.Size() != len(model) {
				t.Fatalf("Step %d: size mismatch: got %d, expected %d", i, q.Size(), len(model))
			}
			if q.Empty() != (len(model) == 0) {
				t.Fatalf("Step %d: empty mismatch: got %v, expected %v", i, q.Empty(), len(model) == 0)
			}
		}
	})
}
