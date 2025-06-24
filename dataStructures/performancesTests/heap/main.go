package main

import (
	"fmt"
	"math/rand/v2"
	"os"
	"runtime/pprof"
	"strings"

	datastructures "archive-tools-monorepo/dataStructures"
)

const (
	targetRuns       = 100
	targetOperations = 1000000
)

func generateNextOp() string {
	op := rand.Int() % 5
	operations := []string{"o", "s", "e", "k"}

	if op >= len(operations) {
		return fmt.Sprintf("p:%x", rand.Float64())
	}

	return operations[op]
}

func main() {
	f, err := os.Create("heap.prof")
	if err != nil {
		panic(err)
	}

	err = pprof.StartCPUProfile(f)
	if err != nil {
		panic(err)
	}

	defer pprof.StopCPUProfile()

	for range targetRuns {
		ourHeap, err := datastructures.NewHeap(func(a, b *string) bool {
			return *a < *b
		})
		if err != nil {
			panic(err)
		}

		for range targetOperations {
			raw := generateNextOp()

			switch {
			case strings.HasPrefix(raw, "p:"):
				valStr := strings.TrimPrefix(raw, "p:")
				err := ourHeap.Push(valStr)
				if err != nil {
					panic(err)
				}
			case raw == "o":
				_, err = ourHeap.Pop()
				if err != nil {
					panic(err)
				}
			case raw == "k":
				_ = ourHeap.Peak()
			case raw == "s":
				_ = ourHeap.Size()
			case raw == "e":
				_ = ourHeap.Empty()
			default:
				continue
			}
		}
	}
}
