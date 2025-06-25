package datastructures_test

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	datastructures "archive-tools-monorepo/dataStructures"
)

func HeapStateMachine[T any](instructions string, parseFN func(string) (T, error), compareFN func(*T, *T) bool) {
	heap, err := datastructures.NewHeap(compareFN)
	if err != nil {
		panic(err)
	}

	// Track expected state
	var model []T

	operations := strings.Split(instructions, ";")

	for i, raw := range operations {
		if raw == "" {
			continue
		}

		switch {
		case strings.HasPrefix(raw, "p:"):
			var val T
			valStr := strings.TrimPrefix(raw, "p:")
			val, err = parseFN(valStr)
			if err != nil {
				continue
			}

			err = heap.Push(val)
			if err != nil {
				panic(err)
			}

			model = append(model, val)
			sort.Slice(model, func(a, b int) bool {
				return compareFN(&model[a], &model[b])
			})

		case raw == "o":
			var expected T
			var result T
			result, err = heap.Pop()
			if err != nil {
				panic(err)
			}

			if !heap.Empty() {
				expected = model[0]
				model = model[1:]
			} else if len(model) != 0 {
				panic(fmt.Sprintf("Step %d: heap state inconsistency - our heap empty but model has %d items", i, len(model)))
			}

			if !reflect.DeepEqual(result, expected) {
				panic(fmt.Sprintf("Step %d: pop heap should return %v value, got %v", i, expected, result))
			}

		case raw == "k":
			sizeBefore := heap.Size()
			peak := heap.Peak()

			if heap.Empty() {
				if peak != nil {
					panic(fmt.Sprintf("Step %d: peak on empty heap should return nil, got %v", i, *peak))
				}
			} else {
				if peak == nil {
					panic(fmt.Sprintf("Step %d: peak on non-empty heap returned nil", i))
				}

				sizeAfter := heap.Size()
				if sizeBefore != sizeAfter {
					panic(fmt.Sprintf("Step %d: peak operation modified heap size", i))
				}
			}

		case raw == "s":
			ourSize := heap.Size()
			modelSize := len(model)

			if ourSize != modelSize {
				panic(fmt.Sprintf("Step %d: size mismatch with model - got %d, expected %d", i, ourSize, modelSize))
			}

		case raw == "e":
			ourEmpty := heap.Empty()
			modelEmpty := len(model) == 0

			if ourEmpty != modelEmpty {
				panic(
					fmt.Sprintf(
						"Step %d: empty state mismatch with model - got %v, expected %v",
						i,
						ourEmpty,
						modelEmpty,
					),
				)
			}

		default:
			continue
		}

		ourSize := heap.Size()
		modelSize := len(model)

		if ourSize != modelSize {
			panic(fmt.Sprintf("Step %d: size invariant violation - our: %d, model: %d", i, ourSize, modelSize))
		}

		ourEmpty := heap.Empty()
		expectedEmpty := (ourSize == 0)
		if ourEmpty != expectedEmpty {
			panic(fmt.Sprintf("Step %d: empty invariant violation - empty: %v, size: %d", i, ourEmpty, ourSize))
		}

		if !heap.Empty() {
			peak := heap.Peak()
			if peak == nil {
				panic(fmt.Sprintf("Step %d: peak returned nil on non-empty heap", i))
			}

			if len(model) > 0 && !reflect.DeepEqual(*peak, model[0]) {
				panic(fmt.Sprintf("Step %d: heap property violation - peak %v != ref min %v", i, *peak, model[0]))
			}
		}
	}

	finalSize := heap.Size()
	finalEmpty := heap.Empty()

	if (finalSize == 0) != finalEmpty {
		panic(fmt.Sprintf("Final state inconsistency: size %d, empty %v", finalSize, finalEmpty))
	}

	var drainedElements []T
	var data T
	for !heap.Empty() {
		data, err = heap.Pop()
		if err != nil {
			panic(err)
		}
		drainedElements = append(drainedElements, data)
	}

	for i := 1; i < len(drainedElements); i++ {
		if compareFN(&drainedElements[i], &drainedElements[i-1]) {
			panic(fmt.Sprintf("Heap property violation in drained elements: %v", drainedElements))
		}
	}
}

func TestHeap_NewHeap_WhenFunctionNotSet_Panic(t *testing.T) {
	var testFn func(*int, *int) bool
	heap, err := datastructures.NewHeap(testFn)

	if err == nil {
		t.Error("Expected panic on new for nil function input, but got none")
	} else if !strings.Contains(fmt.Sprintf("%v", err), "provided function is a nil pointer") {
		t.Errorf("Unexpected panic message: %v", err)
	}

	err = heap.Push(2)

	if err == nil {
		t.Error("Expected panic on push for nil function input, but got none")
	} else if !strings.Contains(fmt.Sprintf("%v", err), "comapre function not set") {
		t.Errorf("Unexpected panic message: %v", err)
	}

	_, err = heap.Pop()

	if err == nil {
		t.Error("Expected panic on pop for nil function input, but got none")
	} else if !strings.Contains(fmt.Sprintf("%v", err), "comapre function not set") {
		t.Errorf("Unexpected panic message: %v", err)
	}
}

func TestHeap_BasicOperations_OK(_ *testing.T) {
	instructions := "p:10;p:30;o;s;p:9;p:8;p:7;k;p:6;p:5;k;p:9;p:40;k;p:39;p:-1;o;o;o;p:1"
	parseFN := strconv.Atoi
	compareFN := func(a, b *int) bool {
		return *a < *b
	}

	HeapStateMachine(instructions, parseFN, compareFN)
}
