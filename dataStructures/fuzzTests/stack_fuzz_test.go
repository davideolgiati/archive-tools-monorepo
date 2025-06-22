package fuzztests

import (
	datastructures "archive-tools-monorepo/dataStructures"
	"strings"
	"testing"
)

func FuzzStack(f *testing.F) {
	testCases := []string{
		"p:1;s;o;e",
		"p:hello;p:world;o;o",
		"s;e",
	}

	for _, testCase := range testCases {
		f.Add(testCase)
	}

	f.Fuzz(func(t *testing.T, tc string) {
		var s datastructures.Stack[string]

		var model []string

		for i, raw := range strings.Split(tc, ";") {
			switch {
			case strings.HasPrefix(raw, "p:"):
				// Push operation
				valStr := strings.TrimPrefix(raw, "p:")
				s.Push(valStr)
				model = append(model, valStr)
			case raw == "o":
				// Pop operation
				val, err := s.Pop()
				if len(model) == 0 {
					if err == nil {
						t.Fatalf("Step %d: expected error on empty pop, got %v", i, val)
					}
				} else {
					expected := model[len(model)-1]
					model = model[:len(model)-1]
					if err != nil {
						t.Fatalf("Step %d: unexpected error: %v", i, err)
					}
					if val != expected {
						t.Fatalf("Step %d: pop mismatch: got %v, expected %v", i, val, expected)
					}
				}
			case raw == "s":
				queueLen := s.Size()
				if queueLen != len(model) {
					t.Fatalf("Step %d: size mismatch: got %v, expected %v", i, queueLen, len(model))
				}
			case raw == "e":
				isEmpty := s.Empty()
				if isEmpty != (len(model) == 0) {
					t.Fatalf("Step %d: size mismatch: got %v, expected %v", i, isEmpty, len(model) == 0)
				}
			default:
				// Unknown op, skip
				continue
			}

			// Invariant checks
			if s.Size() != len(model) {
				t.Fatalf("Step %d: size mismatch: got %d, expected %d", i, s.Size(), len(model))
			}
			if s.Empty() != (len(model) == 0) {
				t.Fatalf("Step %d: empty mismatch: got %v, expected %v", i, s.Empty(), len(model) == 0)
			}
		}
	})
}
