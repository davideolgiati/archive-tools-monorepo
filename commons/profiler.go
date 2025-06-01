package commons

import (
	"fmt"
	"runtime/metrics"
	"sort"
	"sync"
	"time"
	"unsafe"
)

const target_sampling_population = 5000

type Profiler struct {
	memory_used    []uint64
	quit_channel   chan bool
	wg             sync.WaitGroup
	start_time     time.Time
	memory_samples int
}

func (pf *Profiler) size() uint64 {
	return uint64(unsafe.Sizeof(*pf)) + uint64(len(pf.memory_used)*8)
}

func (pf *Profiler) Start() {
	pf.start_time = time.Now()
	pf.quit_channel = make(chan bool)
	pf.memory_used = make([]uint64, target_sampling_population)

	pf.wg.Add(1)

	go func(wg *sync.WaitGroup, quit_channel chan bool) {
		defer wg.Done()

		sample := make([]metrics.Sample, 1)
		sample[0].Name = "/memory/classes/total:bytes"
		index := 0
		pf.memory_samples = 0

		for {
			select {
			case <-quit_channel:
				return
			default:
				metrics.Read(sample)

				if sample[0].Value.Kind() != metrics.KindBad {
					pf.memory_used[index] = (sample[0].Value.Uint64() - pf.size())
					index = (index + 1) % target_sampling_population
					if pf.memory_samples < target_sampling_population {
						pf.memory_samples++
					}
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

	if pf.memory_samples < target_sampling_population {
		pf.memory_used = pf.memory_used[:pf.memory_samples]
	}

	sort.Slice(pf.memory_used, func(i, j int) bool { return pf.memory_used[i] < pf.memory_used[j] })

	p50index := pf.memory_samples / 2
	p90index := (pf.memory_samples * 90) / 100
	p99index := (pf.memory_samples * 99) / 100

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
