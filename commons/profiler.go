package commons

import (
	"fmt"
	"runtime/metrics"
	"sort"
	"sync"
	"time"
)

type Profiler struct {
	memory_used  []uint64
	quit_channel chan bool
	wg sync.WaitGroup
}

func (pf *Profiler) Start() {
	pf.quit_channel = make(chan bool)

	pf.wg.Add(1)

	go func(wg *sync.WaitGroup, quit_channel chan bool) {
		defer wg.Done()

		//tmp_stack := ds.Stack[uint64]{}
		sample := make([]metrics.Sample, 1)
		sample[0].Name = "/memory/classes/total:bytes"

		for {
			select {
			case <-quit_channel:
				return
			default:
				metrics.Read(sample)

				if sample[0].Value.Kind() != metrics.KindBad {
					pf.memory_used = append(pf.memory_used, sample[0].Value.Uint64())
				}

				time.Sleep(300 * time.Millisecond)

				// if pf.memory_used.Size() == 500 {
				// 	count := 1

				// 	for !pf.memory_used.Empty() {
				// 		data := pf.memory_used.Pop()

				// 		if count % 2 == 0 {
				// 			tmp_stack.Push(data)
				// 		}

				// 		count++
				// 	}

				// 	for !tmp_stack.Empty() {
				// 		pf.memory_used.Push(tmp_stack.Pop())
				// 	}
				// }
			}
		}
	}(&pf.wg, pf.quit_channel)
}

func (pf *Profiler) Collect() {
	pf.quit_channel <- true
	pf.wg.Wait()

	// Memory

	sort.Slice(pf.memory_used, func(i, j int) bool { return pf.memory_used[i] < pf.memory_used[j] })

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

