package commons

import (
	"archive-tools-monorepo/commons/ds"
	"fmt"
	"runtime/metrics"
	"sync"
	"time"
)

type Profiler struct {
	memory_used  ds.Heap[uint64]
	quit_channel chan bool
	wg sync.WaitGroup
}

func (pf *Profiler) Start() {
	pf.memory_used = ds.Heap[uint64]{}
	pf.memory_used.Compare_fn(func(a uint64, b uint64) bool {
		return a < b
	})
	pf.quit_channel = make(chan bool)

	pf.wg.Add(1)

	go func(wg *sync.WaitGroup, quit_channel chan bool) {
		defer wg.Done()

		for {
			select {
			case <-quit_channel:
				return
			default:
				//sample := make([]metrics.Sample, 2)
				sample := make([]metrics.Sample, 1)
				sample[0].Name = "/memory/classes/total:bytes"
				//sample[1].Name = "/gc/heap/objects:objects"
				metrics.Read(sample)

				if sample[0].Value.Kind() != metrics.KindBad {
					pf.memory_used.Push(sample[0].Value.Uint64())
				}

				time.Sleep(100 * time.Millisecond)

				if pf.memory_used.Size() == 10000 {
					tmp_stack := ds.Stack[uint64]{}
					count := 1

					for !pf.memory_used.Empty() {
						data := pf.memory_used.Pop()

						if count % 2 == 0 {
							tmp_stack.Push(data)
						}

						count++
					}

					for !tmp_stack.Empty() {
						pf.memory_used.Push(tmp_stack.Pop())
					}
				}
			}
		}
	}(&pf.wg, pf.quit_channel)
}

func (pf *Profiler) Collect() {
	pf.quit_channel <- true
	pf.wg.Wait()

	// Memory

	memory := make([]uint64, 0, pf.memory_used.Size())

	for !pf.memory_used.Empty() {
		memory = append(memory, pf.memory_used.Pop())
	}

	p50index := len(memory) / 2
	p90index := (len(memory) * 9) / 10
	p99index := (len(memory) * 99) / 100

	p50 := Format_file_size(int64(memory[p50index]))
	p90 := Format_file_size(int64(memory[p90index]))
	p99 := Format_file_size(int64(memory[p99index]))

	fmt.Printf(
		"Memery usage:\n\tp50: %d %s\n\tp90: %d %s\n\tp99: %d %s\n",
		p50.Value, *p50.Unit,
		p90.Value, *p90.Unit,
		p99.Value, *p99.Unit,
	)
}

