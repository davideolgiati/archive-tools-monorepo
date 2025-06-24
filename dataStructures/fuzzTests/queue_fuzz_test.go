package datastructures_fuzz_test

import (
	"strings"
	"testing"

	datastructures "archive-tools-monorepo/dataStructures"
)

func FuzzQueue(f *testing.F) {
	testCases := []string{
		"p:1;s;o;e",
		"p:hello;p:world;o;o",
		"s;e",
	}

	for _, testCase := range testCases {
		f.Add(testCase)
	}

	f.Fuzz(func(t *testing.T, tc string) {
		var q datastructures.Queue[string]
		q.Init()

		var model []string

		for i, raw := range strings.Split(tc, ";") {
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
