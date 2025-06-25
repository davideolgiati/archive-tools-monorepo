package datastructures_fuzz_test

import (
	"container/heap"
	"sort"
	"strconv"
	"strings"
	"testing"

	datastructures "archive-tools-monorepo/dataStructures"
)

// Reference implementation using Go's standard library for comparison.
type IntHeap []int

func (h *IntHeap) Len() int           { return len(*h) }
func (h *IntHeap) Less(i, j int) bool { return (*h)[i] < (*h)[j] }
func (h *IntHeap) Swap(i, j int)      { (*h)[i], (*h)[j] = (*h)[j], (*h)[i] }

func (h *IntHeap) Push(x interface{}) {
	*h = append(*h, x.(int))
}

func (h *IntHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func FuzzHeap(f *testing.F) {
	// Seed with various operation patterns
	testCases := []string{
		// Basic operations
		"p:1;p:2;p:3;o;o;o",
		"p:5;p:1;p:3;o;p:2;o;o;o",
		// Size and empty checks
		"p:10;s;e;o;s;e",
		// Peak operations
		"p:1;k;p:0;k;o;k",
		// Large numbers to test overflow scenarios
		"p:2147483647;p:-2147483648;o;o",
		// Duplicate values
		"p:5;p:5;p:5;o;o;o",
		// Mixed operations
		"p:3;p:1;k;p:4;o;k;p:2;o;o;o",
		// Empty operations (should handle gracefully)
		"o;k;s;e",
		// Stress test with many operations
		"p:1;p:2;p:3;p:4;p:5;o;o;p:6;p:7;o;o;o;o;o",
	}

	for _, testCase := range testCases {
		f.Add(testCase)
	}

	f.Fuzz(func(t *testing.T, tc string) {
		// Create our heap with integer comparison (min heap)
		ourHeap, err := datastructures.NewHeap(func(a, b *int) bool {
			return *a < *b
		})
		if err != nil {
			panic(err)
		}

		// Reference heap for comparison
		var refHeap IntHeap
		heap.Init(&refHeap)

		// Track expected state
		var model []int

		operations := strings.Split(tc, ";")

		for i, raw := range operations {
			// Skip empty operations
			if raw == "" {
				continue
			}

			switch {
			case strings.HasPrefix(raw, "p:"):
				// Push operation - extract integer value
				var val int
				valStr := strings.TrimPrefix(raw, "p:")
				val, err = strconv.Atoi(valStr)
				if err != nil {
					// Skip invalid integers to avoid test failures on random input
					continue
				}

				// Security check: prevent extremely large slice allocations
				if ourHeap.Size() > 10000 {
					continue
				}

				err = ourHeap.Push(val)
				if err != nil {
					panic(err)
				}

				heap.Push(&refHeap, val)
				model = append(model, val)
				sort.Ints(model) // Keep model sorted for min-heap comparison

			case raw == "o":
				var expected int
				var result int

				if !ourHeap.Empty() {
					expected = heap.Pop(&refHeap).(int)
				}

				result, err = ourHeap.Pop()
				if err != nil {
					panic(err)
				}

				if ourHeap.Empty() && len(refHeap) != 0 {
					t.Fatalf("Step %d: heap state inconsistency - our heap empty but ref heap has %d items",
						i, len(refHeap))
				}

				if result != expected {
					t.Fatalf("Step %d: pop mismatch - got %v, expected %v", i, result, expected)
				}

				if len(model) > 0 && model[0] != result {
					t.Fatalf("Step %d: model inconsistency - expected min %v, got %v",
						i, model[0], result)
				}

				if len(model) > 0 {
					model = model[1:]
				}

			case raw == "k":
				sizeBefore := ourHeap.Size()
				peak := ourHeap.Peak()
				sizeAfter := ourHeap.Size()

				if sizeBefore != sizeAfter {
					t.Fatalf("Step %d: peak operation modified heap size", i)
				}

				if ourHeap.Empty() && peak != nil {
					t.Fatalf("Step %d: peak on empty heap should return nil, got %v", i, *peak)
				} else if !ourHeap.Empty() && peak == nil {
					t.Fatalf("Step %d: peak on non-empty heap returned nil", i)
				}

				if ourHeap.Empty() {
					continue
				}

				expectedPeak := refHeap[0]
				if *peak != expectedPeak {
					t.Fatalf("Step %d: peak mismatch - got %v, expected %v",
						i, *peak, expectedPeak)
				}

			case raw == "s":
				ourSize := ourHeap.Size()
				refSize := len(refHeap)
				modelSize := len(model)

				if ourSize != refSize {
					t.Fatalf("Step %d: size mismatch with reference - got %d, expected %d",
						i, ourSize, refSize)
				}

				if ourSize != modelSize {
					t.Fatalf("Step %d: size mismatch with model - got %d, expected %d",
						i, ourSize, modelSize)
				}

			case raw == "e":
				ourEmpty := ourHeap.Empty()
				refEmpty := len(refHeap) == 0
				modelEmpty := len(model) == 0

				if ourEmpty != refEmpty {
					t.Fatalf("Step %d: empty state mismatch with reference - got %v, expected %v",
						i, ourEmpty, refEmpty)
				}

				if ourEmpty != modelEmpty {
					t.Fatalf("Step %d: empty state mismatch with model - got %v, expected %v",
						i, ourEmpty, modelEmpty)
				}

			default:
				continue
			}

			// Invariant checks after each operation
			ourSize := ourHeap.Size()
			refSize := len(refHeap)
			modelSize := len(model)

			// Size consistency
			if ourSize != refSize || ourSize != modelSize {
				t.Fatalf("Step %d: size invariant violation - our: %d, ref: %d, model: %d",
					i, ourSize, refSize, modelSize)
			}

			// Empty state consistency
			ourEmpty := ourHeap.Empty()
			expectedEmpty := (ourSize == 0)
			if ourEmpty != expectedEmpty {
				t.Fatalf("Step %d: empty invariant violation - empty: %v, size: %d",
					i, ourEmpty, ourSize)
			}

			// Heap property check (if not empty)
			if !ourHeap.Empty() {
				peak := ourHeap.Peak()
				if peak == nil {
					t.Fatalf("Step %d: peak returned nil on non-empty heap", i)
				}

				// Peak should match reference heap's minimum
				if len(refHeap) > 0 && *peak != refHeap[0] {
					t.Fatalf("Step %d: heap property violation - peak %v != ref min %v",
						i, *peak, refHeap[0])
				}
			}

			// Security check: prevent resource exhaustion
			if ourSize > 10000 {
				t.Fatalf("Step %d: heap size exceeded safety limit: %d", i, ourSize)
			}
		}

		// Final verification - ensure heap is in valid state
		finalSize := ourHeap.Size()
		finalEmpty := ourHeap.Empty()

		if (finalSize == 0) != finalEmpty {
			t.Fatalf("Final state inconsistency: size %d, empty %v", finalSize, finalEmpty)
		}

		// Drain remaining elements and verify they come out in sorted order
		var drainedElements []int
		var data int
		for !ourHeap.Empty() {
			data, err = ourHeap.Pop()
			if err != nil {
				panic(err)
			}
			drainedElements = append(drainedElements, data)
		}

		// Verify sorted order (min heap property)
		for i := 1; i < len(drainedElements); i++ {
			if drainedElements[i] < drainedElements[i-1] {
				t.Fatalf("Heap property violation in drained elements: %v", drainedElements)
			}
		}
	})
}
