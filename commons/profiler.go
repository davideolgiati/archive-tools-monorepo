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
	wg           sync.WaitGroup
	start_time   time.Time
}

func (pf *Profiler) size() uint64 {
	return uint64(1000*64) + 8 + 16
}

func (pf *Profiler) Start() {
	pf.start_time = time.Now()
	pf.quit_channel = make(chan bool)
	pf.memory_used = make([]uint64, 1000)

	pf.wg.Add(1)

	go func(wg *sync.WaitGroup, quit_channel chan bool) {
		defer wg.Done()

		sample := make([]metrics.Sample, 1)
		sample[0].Name = "/memory/classes/total:bytes"
		index := 0

		for {
			select {
			case <-quit_channel:
				return
			default:
				metrics.Read(sample)

				if sample[0].Value.Kind() != metrics.KindBad {
					pf.memory_used[index] = (sample[0].Value.Uint64() - pf.size())
					index = (index + 1) % 1000
				}
			}
		}
	}(&pf.wg, pf.quit_channel)
}

func (pf *Profiler) Collect() {
	pf.quit_channel <- true

	fmt.Printf("Duration : %v\n", time.Since(pf.start_time))

	pf.wg.Wait()

	// Memory

	sort.Slice(pf.memory_used, func(i, j int) bool { return pf.memory_used[i] < pf.memory_used[j] })

	p50index := 500
	p90index := 900
	p99index := 990

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
