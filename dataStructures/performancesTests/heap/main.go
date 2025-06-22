package main

import (
	"archive-tools-monorepo/dataStructures"
	"fmt"
	"math/rand/v2"
	"os"
	"runtime/pprof"
	"strings"
)

func generateNextOp() string {
	op := rand.Int() % 5

	switch {
	case op == 0:
		return "o"
	case op == 1:
		return "s"
	case op == 2:
		return "e"
	case op == 3:
		return "k"
	default:
		return fmt.Sprintf("p:%x", rand.Float64())
	}
}

func main() {
	f, err := os.Create("heap.prof")
	if err != nil {
		panic(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	for x := 0; x < 100; x++ {
		ourHeap, err := dataStructures.NewHeap(func(a, b *string) bool {
			return *a < *b
		})

		if err != nil {
			panic(err)
		}

		for i := 0; i < 10000000; i++ {
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
