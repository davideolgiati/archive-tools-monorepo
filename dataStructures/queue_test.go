package datastructures_test

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	datastructures "archive-tools-monorepo/dataStructures"
)

func QueueStateMachine[T any](instructions string, parseFN func(string) (T, error), compareFN func(*T, *T) bool) {
	queue := datastructures.Queue[T]{}
	queue.Init()

	// Track expected state
	var model []T

	operations := strings.Split(instructions, ";")

	for i, raw := range operations {
		if raw == "" {
			continue
		}

		switch {
		case strings.HasPrefix(raw, "p:"):
			valStr := strings.TrimPrefix(raw, "p:")
			val, err := parseFN(valStr)
			if err != nil {
				continue
			}

			queue.Push(val)

			model = append(model, val)
		case raw == "o":
			var expected T
			initialQueueState := queue.Empty() 

			if !queue.Empty() {
				expected = model[0]
				model = model[1:]
			}

			result, err := queue.Pop()

			if !initialQueueState && err != nil {
				panic(fmt.Sprintf("Step %d: queue state inconsistency - expected nil got error: %v", i, err))
			}

			if initialQueueState && err == nil {
				panic(fmt.Sprintf("Step %d: queue state inconsistency - expected error got nil", i))
			}

			if queue.Empty() && len(model) != 0 {
				panic(fmt.Sprintf("Step %d: queue state inconsistency - our queue empty but model has %d items", i, len(model)))
			}

			if !reflect.DeepEqual(result, expected) {
				panic(fmt.Sprintf("Step %d: pop queue should return %v value, got %v", i, expected, result))
			}

		case raw == "k":
			sizeBefore := queue.Size()
			peak := queue.Peak()

			if queue.Empty() {
				if peak != nil {
					panic(fmt.Sprintf("Step %d: peak on empty queue should return nil, got %v", i, *peak))
				}
			} else {
				if peak == nil {
					panic(fmt.Sprintf("Step %d: peak on non-empty queue returned nil", i))
				}

				sizeAfter := queue.Size()
				if sizeBefore != sizeAfter {
					panic(fmt.Sprintf("Step %d: peak operation modified queue size", i))
				}
			}

		case raw == "s":
			ourSize := queue.Size()
			modelSize := len(model)

			if ourSize != modelSize {
				panic(fmt.Sprintf("Step %d: size mismatch with model - got %d, expected %d", i, ourSize, modelSize))
			}

		case raw == "e":
			ourEmpty := queue.Empty()
			modelEmpty := len(model) == 0

			if ourEmpty != modelEmpty {
				panic(fmt.Sprintf("Step %d: empty state mismatch with model - got %v, expected %v", i, ourEmpty, modelEmpty))
			}

		default:
			continue
		}

		ourSize := queue.Size()
		modelSize := len(model)

		if ourSize != modelSize {
			panic(fmt.Sprintf("Step %d: size invariant violation - our: %d, model: %d", i, ourSize, modelSize))
		}

		ourEmpty := queue.Empty()
		expectedEmpty := (ourSize == 0)
		if ourEmpty != expectedEmpty {
			panic(fmt.Sprintf("Step %d: empty invariant violation - empty: %v, size: %d", i, ourEmpty, ourSize))
		}

		if !queue.Empty() {
			peak := queue.Peak()
			if peak == nil {
				panic(fmt.Sprintf("Step %d: peak returned nil on non-empty queue", i))
			}

			if len(model) > 0 && !reflect.DeepEqual(*peak, model[0]) {
				panic(fmt.Sprintf("Step %d: queue property violation - peak %v != ref min %v", i, *peak, model[0]))
			}
		}
	}

	finalSize := queue.Size()
	finalEmpty := queue.Empty()

	if (finalSize == 0) != finalEmpty {
		panic(fmt.Sprintf("Final state inconsistency: size %d, empty %v", finalSize, finalEmpty))
	}

	var drainedElements []T
	for !queue.Empty() {
		data, err := queue.Pop()
		if err != nil {
			panic(err)
		}
		drainedElements = append(drainedElements, data)
	}

	for i := 1; i < len(drainedElements); i++ {
		if !reflect.DeepEqual(drainedElements[i], model[i]) {
			panic(fmt.Sprintf("Queue property violation in drained elements: %v", drainedElements))
		}
	}
}

func TestQueue_BasicOperations_OK(_ *testing.T) {
	instructions := "k;o;p:10;o;p:30;o;s;p:9;p:8;p:7;k;p:6;p:5;k;p:9;p:40;k;p:39;p:-1;o;o;o;p:1"
	parseFN := strconv.Atoi
	compareFN := func(a, b *int) bool {
		return *a < *b
	}

	QueueStateMachine(instructions, parseFN, compareFN)
}
