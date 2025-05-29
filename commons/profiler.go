package commons

import (
	"cmp"
	"fmt"
	"runtime/metrics"
	"slices"
	"sync"
	"time"
)

type Profiler struct {
	memory_used  []uint64
	heap_obj_count  []uint64
	quit_channel chan bool
	wg sync.WaitGroup
}

func (pf *Profiler) Start() {
	pf.quit_channel = make(chan bool)

	pf.wg.Add(1)

	go func(wg *sync.WaitGroup, quit_channel chan bool) {
		defer wg.Done()

		for {
			select {
			case <-quit_channel:
				return
			default:
				sample := make([]metrics.Sample, 2)
				sample[0].Name = "/memory/classes/total:bytes"
				sample[1].Name = "/gc/heap/objects:objects"
				metrics.Read(sample)

				if sample[0].Value.Kind() != metrics.KindBad {
					pf.memory_used = append(pf.memory_used, sample[0].Value.Uint64())
				}

				if sample[1].Value.Kind() != metrics.KindBad {
					pf.heap_obj_count = append(pf.heap_obj_count, sample[1].Value.Uint64())
				}

				time.Sleep(1 * time.Millisecond)
			}
		}
	}(&pf.wg, pf.quit_channel)
}

func (pf *Profiler) Collect() {
	pf.quit_channel <- true
	pf.wg.Wait()

	// Memory

	slices.SortFunc(
		pf.memory_used,
	        func(a, b uint64) int {
			return cmp.Compare(a, b)
		},
	)

	p50index := len(pf.memory_used) / 2
	p90index := (len(pf.memory_used) * 9) / 10
	p99index := (len(pf.memory_used) * 99) / 100

	p50 := Format_file_size(int64(pf.memory_used[p50index]))
	p90 := Format_file_size(int64(pf.memory_used[p90index]))
	p99 := Format_file_size(int64(pf.memory_used[p99index]))

	fmt.Printf(
		"Memery usage:\n\tp50: %d %s\n\tp90: %d %s\n\tp99: %d %s\n",
		p50.Value, *p50.Unit,
		p90.Value, *p90.Unit,
		p99.Value, *p99.Unit,
	)
}

