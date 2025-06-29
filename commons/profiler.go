package commons

import (
	"fmt"
	"runtime/metrics"
	"sort"
	"sync"
	"time"
	"unsafe"
)

const targetSamplingPopulation = 5000

type Profiler struct {
	startTime     time.Time
	quitChannel   chan bool
	memoryUsed    []uint64
	wg            sync.WaitGroup
	memorySamples int
}

func (pf *Profiler) size() uint64 {
	var profilerNominalSize interface{} = unsafe.Sizeof(*pf)
	var memorySamplesSize interface{} = len(pf.memoryUsed)

	var totalMemoryUsed uint64

	temp, ok := profilerNominalSize.(uint64)

	if ok {
		totalMemoryUsed += temp
	} else {
		panic("error while converting profilerNominalSize to uint64")
	}

	temp, ok = memorySamplesSize.(uint64)

	if ok {
		totalMemoryUsed += (temp * 8)
	} else {
		panic("error while converting totalMemoryUsed to uint64")
	}

	return totalMemoryUsed
}

func (pf *Profiler) Start() {
	pf.startTime = time.Now()
	pf.quitChannel = make(chan bool)
	pf.memoryUsed = make([]uint64, targetSamplingPopulation)

	pf.wg.Add(1)

	go func(wg *sync.WaitGroup, quitChannel chan bool) {
		defer wg.Done()

		sample := make([]metrics.Sample, 1)
		sample[0].Name = "/memory/classes/total:bytes"
		index := 0
		pf.memorySamples = 0

		for {
			select {
			case <-quitChannel:
				return
			default:
				metrics.Read(sample)

				if sample[0].Value.Kind() != metrics.KindBad {
					pf.memoryUsed[index] = (sample[0].Value.Uint64() - pf.size())
					index = (index + 1) % targetSamplingPopulation
					if pf.memorySamples < targetSamplingPopulation {
						pf.memorySamples++
					}
				}
			}
		}
	}(&pf.wg, pf.quitChannel)
}

func (pf *Profiler) Collect() {
	pf.quitChannel <- true

	fmt.Printf("Duration : %v\n", time.Since(pf.startTime))

	pf.wg.Wait()

	// Memory

	if pf.memorySamples < targetSamplingPopulation {
		pf.memoryUsed = pf.memoryUsed[:pf.memorySamples]
	}

	sort.Slice(pf.memoryUsed, func(i, j int) bool { return pf.memoryUsed[i] < pf.memoryUsed[j] })

	p50index := pf.memorySamples / 2
	p90index := (pf.memorySamples * 90) / 100
	p99index := (pf.memorySamples * 99) / 100

	var p50value interface{} = pf.memoryUsed[p50index]
	p50memoryUsed, ok := p50value.(int64)

	if !ok {
		panic("can't safely cast p50value memrory used")
	}

	var p90value interface{} = pf.memoryUsed[p90index]
	p90memoryUsed, ok := p90value.(int64)

	if !ok {
		panic("can't safely cast p90value memrory used")
	}

	var p99value interface{} = pf.memoryUsed[p99index]
	p99memoryUsed, ok := p99value.(int64)

	if !ok {
		panic("can't safely cast p99value memrory used")
	}

	p50, err := FormatFileSize(p50memoryUsed)
	if err != nil {
		panic(err)
	}

	p90, err := FormatFileSize(p90memoryUsed)
	if err != nil {
		panic(err)
	}

	p99, err := FormatFileSize(p99memoryUsed)
	if err != nil {
		panic(err)
	}

	fmt.Printf(
		"Memery usage:\n\tp50: %d %s\n\tp90: %d %s\n\tp99: %d %s\n",
		p50.Value, *p50.Unit,
		p90.Value, *p90.Unit,
		p99.Value, *p99.Unit,
	)
}
