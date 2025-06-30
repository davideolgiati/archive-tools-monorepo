package datastructures_test

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	datastructures "archive-tools-monorepo/dataStructures"
)

func pushS[T any](value T, stack *datastructures.Stack[T], model *[]T) {
	stack.Push(value)
	*model = append(*model, value)
}

func popS[T any](stack *datastructures.Stack[T], model *[]T, step int) {
	var expected T
	initialstackState := stack.Empty()

	if !stack.Empty() {
		expected = (*model)[len(*model)-1]
		*model = (*model)[:len(*model)-1]
	}

	result, err := stack.Pop()

	if !initialstackState && err != nil {
		panic(fmt.Sprintf("Step %d: stack state inconsistency - expected nil got error: %v", step, err))
	}

	if initialstackState && err == nil {
		panic(fmt.Sprintf("Step %d: stack state inconsistency - expected error got nil", step))
	}

	if stack.Empty() && len(*model) != 0 {
		panic(
			fmt.Sprintf(
				"Step %d: stack state inconsistency - our stack empty but model has %d items",
				step,
				len(*model),
			),
		)
	}

	if !reflect.DeepEqual(result, expected) {
		panic(fmt.Sprintf("Step %d: pop stack should return %v value, got %v", step, expected, result))
	}
}

func peakS[T any](stack *datastructures.Stack[T], model *[]T, step int) {
	var expected T
	sizeBefore := stack.Size()
	peak := stack.Peak()

	if stack.Empty() {
		if peak != nil {
			panic(fmt.Sprintf("Step %d: peak on empty stack should return nil, got %v", step, *peak))
		}
	} else {
		if peak == nil {
			panic(fmt.Sprintf("Step %d: peak on non-empty stack returned nil", step))
		}

		sizeAfter := stack.Size()
		if sizeBefore != sizeAfter {
			panic(fmt.Sprintf("Step %d: peak operation modified stack size", step))
		}

		expected = (*model)[len(*model)-1]

		if !reflect.DeepEqual(*peak, expected) {
			panic(fmt.Sprintf("Step %d: peak value inconsistency - expected %v, got %v", step, expected, *peak))
		}
	}
}

func stackStateMachine[T any](instructions string, parseFN func(string) (T, error)) {
	stack := datastructures.Stack[T]{}

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
			pushS(val, &stack, &model)

		case raw == "o":
			popS(&stack, &model, i)

		case raw == "k":
			peakS(&stack, &model, i)

		case raw == "s":
			ourSize := stack.Size()
			modelSize := len(model)

			if ourSize != modelSize {
				panic(fmt.Sprintf("Step %d: size mismatch with model - got %d, expected %d", i, ourSize, modelSize))
			}

		case raw == "e":
			ourEmpty := stack.Empty()
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

		ourSize := stack.Size()
		modelSize := len(model)

		if ourSize != modelSize {
			panic(fmt.Sprintf("Step %d: size invariant violation - our: %d, model: %d", i, ourSize, modelSize))
		}

		ourEmpty := stack.Empty()
		expectedEmpty := (ourSize == 0)
		if ourEmpty != expectedEmpty {
			panic(fmt.Sprintf("Step %d: empty invariant violation - empty: %v, size: %d", i, ourEmpty, ourSize))
		}

		if !stack.Empty() {
			peak := stack.Peak()
			if peak == nil {
				panic(fmt.Sprintf("Step %d: peak returned nil on non-empty stack", i))
			}

			if len(model) > 0 && !reflect.DeepEqual(*peak, model[len(model)-1]) {
				panic(fmt.Sprintf("Step %d: stack property violation - peak %v != ref min %v", i, *peak, model[0]))
			}
		}
	}

	finalSize := stack.Size()
	finalEmpty := stack.Empty()

	if (finalSize == 0) != finalEmpty {
		panic(fmt.Sprintf("Final state inconsistency: size %d, empty %v", finalSize, finalEmpty))
	}

	var drainedElements []T
	for !stack.Empty() {
		data, err := stack.Pop()
		if err != nil {
			panic(err)
		}
		drainedElements = append(drainedElements, data)
	}

	for i := 1; i < len(drainedElements); i++ {
		if !reflect.DeepEqual(drainedElements[i], model[len(model)-1-i]) {
			panic(fmt.Sprintf("stack property violation in drained elements: %v", drainedElements))
		}
	}
}

func TestStack_BasicOperations_OK(_ *testing.T) {
	instructions := "k;o;p:10;o;p:30;o;s;p:9;p:8;p:7;k;p:6;p:5;k;p:9;p:40;k;p:39;p:-1;o;o;o;p:1"
	parseFN := strconv.Atoi

	stackStateMachine(instructions, parseFN)
}
