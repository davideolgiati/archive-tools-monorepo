package main

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	"os"
	"runtime/pprof"
	"strings"

	datastructures "archive-tools-monorepo/dataStructures"
)

const (
	targetRuns       = 100
	targetOperations = 1000000
)

func randomFloat64() float64 {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}

	return float64(binary.LittleEndian.Uint64(b[:])) / (1 << 64)
}

func generateNextOp() string {
	operations := []string{"o", "s", "e", "k"}
	n, err := rand.Int(rand.Reader, big.NewInt(5))
	if err != nil {
		panic(err)
	}

	if n.Int64() >= int64(len(operations)) {
		return fmt.Sprintf("p:%x", randomFloat64())
	}

	return operations[n.Int64()]
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
		var ourHeap *datastructures.Heap[string]
		ourHeap, err = datastructures.NewHeap(func(a, b *string) bool {
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
				err = ourHeap.Push(valStr)
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
