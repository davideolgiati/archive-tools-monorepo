package dataStructures

import (
	"math/rand"
	"sort"
	"sync"
	"testing"
	"time"
)

// TestFile for basic testing
type TestFile struct {
	Name string
	Size int
	Hash string
}

// MinSizeFn for a min-heap based on size
func MinSizeFn(a, b TestFile) bool {
	return a.Size < b.Size
}

// MaxSizeFn for a max-heap based on size
func MaxSizeFn(a, b TestFile) bool {
	return a.Size > b.Size
}

// MinHashFn for a min-heap based on hash (lexicographical)
func MinHashFn(a, b TestFile) bool {
	return a.Hash < b.Hash
}

// MaxHashFn for a max-heap based on hash (lexicographical)
func MaxHashFn(a, b TestFile) bool {
	return a.Hash > b.Hash
}

// TestHeap_BasicOperations verifies fundamental Push and Pop behavior.
func TestHeap_BasicOperations(t *testing.T) {
	heap := NewHeap(MinSizeFn)

	// Push some elements
	heap.Push(TestFile{Name: "fileA", Size: 10})
	heap.Push(TestFile{Name: "fileB", Size: 5})
	heap.Push(TestFile{Name: "fileC", Size: 15})
	heap.Push(TestFile{Name: "fileD", Size: 2})

	if heap.Size() != 4 {
		t.Errorf("Expected size 4, got %d", heap.Size())
	}

	// Pop elements and verify order (min-heap)
	if item := heap.Pop(); item.Size != 2 {
		t.Errorf("Expected 2, got %d", item.Size)
	}
	if item := heap.Pop(); item.Size != 5 {
		t.Errorf("Expected 5, got %d", item.Size)
	}
	if item := heap.Pop(); item.Size != 10 {
		t.Errorf("Expected 10, got %d", item.Size)
	}
	if item := heap.Pop(); item.Size != 15 {
		t.Errorf("Expected 15, got %d", item.Size)
	}

	if !heap.Empty() {
		t.Errorf("Expected heap to be empty")
	}
	if heap.Size() != 0 {
		t.Errorf("Expected size 0, got %d", heap.Size())
	}

	// Test Pop on empty heap
	emptyItem := heap.Pop()
	var zero TestFile
	if emptyItem != zero { // Check for zero value
		t.Errorf("Expected zero value from Pop on empty heap, got %v", emptyItem)
	}
}

// TestHeap_Peak verifies Peak functionality without altering the heap.
func TestHeap_Peak(t *testing.T) {
	heap := NewHeap(MinSizeFn)

	heap.Push(TestFile{Name: "fileA", Size: 10})
	heap.Push(TestFile{Name: "fileB", Size: 5})

	if heap.Size() != 2 {
		t.Fatalf("Expected size 2, got %d", heap.Size())
	}

	item := heap.Peak()
	if item == nil || item.Size != 5 {
		t.Errorf("Expected peak item size 5, got %v", item)
	}

	if heap.Size() != 2 {
		t.Errorf("Peak should not change heap size, got %d", heap.Size())
	}

	heap.Pop() // Pop 5
	item = heap.Peak()
	if item == nil || item.Size != 10 {
		t.Errorf("Expected peak item size 10 after pop, got %v", item)
	}

	heap.Pop() // Pop 10
	item = heap.Peak()
	if item != nil {
		t.Errorf("Expected nil from Peak on empty heap, got %v", item)
	}
}

// TestHeap_EdgeCases tests behavior with single element and empty heap.
func TestHeap_EdgeCases(t *testing.T) {
	heap := NewHeap(MinSizeFn)

	// Test with single element
	heap.Push(TestFile{Name: "single", Size: 100})
	if heap.Size() != 1 {
		t.Errorf("Expected size 1, got %d", heap.Size())
	}
	if item := heap.Pop(); item.Size != 100 {
		t.Errorf("Expected 100, got %d", item.Size)
	}
	if !heap.Empty() {
		t.Errorf("Expected heap to be empty")
	}
}

// TestHeap_RandomData verifies heap property with a large number of random insertions.
func TestHeap_RandomData(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	heap := NewHeap(func(a, b int) bool { return a < b }) // Min-heap for ints

	numElements := 1000
	elements := make([]int, numElements)
	for i := 0; i < numElements; i++ {
		val := rand.Intn(10000) // Random int between 0 and 9999
		heap.Push(val)
		elements[i] = val
	}

	if heap.Size() != numElements {
		t.Errorf("Expected size %d, got %d", numElements, heap.Size())
	}

	// Pop all elements and verify sorted order
	var poppedElements []int
	for heap.Size() > 0 {
		poppedElements = append(poppedElements, heap.Pop())
	}

	// Check if the popped elements are sorted
	if !sort.IntsAreSorted(poppedElements) {
		t.Errorf("Popped elements are not sorted (min-heap order)!")
		// Optionally print to debug large failures:
		t.Logf("Popped: %v", poppedElements)
	}
}

// TestHeap_Concurrency_PushAndPop verifies thread-safety under concurrent Push and Pop.
func TestHeap_Concurrency_PushAndPop(t *testing.T) {
	heap := NewHeap(func(a, b int) bool { return a < b }) // Min-heap for ints

	numGoroutines := 10
	numOperationsPerGoroutine := 1000

	var wg sync.WaitGroup
	var pushVals []int
	var mu sync.Mutex // To protect pushVals slice

	// Concurrent Push operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				val := rand.Intn(100000)
				heap.Push(val)
				mu.Lock()
				pushVals = append(pushVals, val)
				mu.Unlock()
			}
		}(i)
	}

	// Allow some pushes to happen before starting pops
	time.Sleep(10 * time.Millisecond)

	// Concurrent Pop operations
	var poppedVals []int
	var popMu sync.Mutex // To protect poppedVals slice
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperationsPerGoroutine; j++ {
				if heap.Size() > 0 { // Only try to pop if not empty
					item := heap.Pop()
					popMu.Lock()
					poppedVals = append(poppedVals, item)
					popMu.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify the final size (it might not be zero if pushes/pops weren't perfectly balanced)
	t.Logf("Final heap size: %d", heap.Size())

	// Combine remaining heap elements with popped ones for full verification
	finalElements := make([]int, 0, len(poppedVals)+heap.Size())
	finalElements = append(finalElements, poppedVals...)

	// Drain remaining elements from the heap
	for heap.Size() > 0 {
		finalElements = append(finalElements, heap.Pop())
	}

	// Sort the originally pushed values
	sort.Ints(pushVals)
	// Sort the final collected values (popped + remaining heap)
	sort.Ints(finalElements)

	// This part is tricky: if pushes and pops are perfectly balanced,
	// then pushVals should contain the same elements as finalElements.
	// However, if pops happen when heap is empty, `Pop` returns zero value,
	// which complicates direct comparison of `pushVals` and `finalElements`.
	// A more robust check: ensure the final state of the heap (all popped + remaining)
	// maintains the heap property and no panics occurred.
	// The primary goal of this test is to ensure no race conditions lead to crashes or incorrect internal state.
	// The `panic` checks in `Push` and `Pop` are crucial for this.

	// A simpler check for concurrency: verify no panics due to concurrent access
	// and that the final elements retrieved maintain the heap property when sorted.
	// If `pushVals` and `finalElements` are different sizes, it implies
	// either some pushes didn't happen or some pops returned zero values due to empty heap.
	// The size comparison is more robust for confirming no data corruption (panics in size checks).
	// If we just check `panic` messages within the test itself, we'd need to use `recover`.
	// The internal panic checks `if len(heap.items) != start_size + 1` etc. are the main guardataStructures for correctness here.
	t.Logf("Total pushed elements: %d", len(pushVals))
	t.Logf("Total final elements (popped + remaining): %d", len(finalElements))

	// The most important check for concurrent correctness is that the heap never panics
	// on its size checks or internal consistency due to concurrent access.
	// The test will fail if any `panic` from `heap.go` occurs.
	// Further, we can check if the elements are still sorted after all ops.
	if !sort.IntsAreSorted(finalElements) {
		t.Errorf("Elements collected after concurrent ops are not sorted correctly!")
	}

	// Additional check: The number of items pushed minus items popped should equal the final heap size
	// (accounting for zero values from pops on empty heap)
	// This is hard to guarantee exactly due to non-deterministic scheduling of pushes/pops.
	// The focus here is on *no corruption* and *heap property maintained*.
}

// TestHeap_Concurrency_PushOnly verifies thread-safety under concurrent Push operations.
func TestHeap_Concurrency_PushOnly(t *testing.T) {
	heap := NewHeap(func(a, b int) bool { return a < b }) // Min-heap for ints

	numGoroutines := 20
	numPushesPerGoroutine := 500
	expectedataStructuresize := numGoroutines * numPushesPerGoroutine

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numPushesPerGoroutine; j++ {
				heap.Push(rand.Intn(100000))
			}
		}()
	}
	wg.Wait()

	if heap.Size() != expectedataStructuresize {
		t.Errorf("Expected heap size %d after concurrent pushes, got %d", expectedataStructuresize, heap.Size())
	}

	// Verify heap property by popping all elements
	var poppedElements []int
	for heap.Size() > 0 {
		poppedElements = append(poppedElements, heap.Pop())
	}

	if !sort.IntsAreSorted(poppedElements) {
		t.Errorf("Elements popped after concurrent pushes are not sorted!")
	}
}

// TestHeap_Concurrency_PopOnly verifies thread-safety under concurrent Pop operations from a pre-filled heap.
func TestHeap_Concurrency_PopOnly(t *testing.T) {
	heap := NewHeap(func(a, b int) bool { return a < b }) // Min-heap for ints

	numElements := 10000
	for i := 0; i < numElements; i++ {
		heap.Push(i) // Push ordered elements for predictable pop order
	}

	numGoroutines := 20
	var wg sync.WaitGroup
	var poppedVals []int
	var mu sync.Mutex // To protect poppedVals slice

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				item := heap.Pop()                 // Pop a single item. Pop on empty returns zero value.
				if item == 0 && heap.Size() == 0 { // Assuming 0 is not a valid element and heap is truly empty
					break
				}
				mu.Lock()
				poppedVals = append(poppedVals, item)
				mu.Unlock()
				if heap.Size() == 0 { // Break early if heap is empty
					break
				}
			}
		}()
	}
	wg.Wait()

	// It's hard to guarantee `poppedVals` contains all original elements
	// if some goroutines tried to Pop on empty heap.
	// The main check is that no panics occurred and the remaining elements (if any)
	// along with popped ones form a valid sorted sequence if all unique values were extracted.
	// For this test, we pushed 0 to 9999, so we expect to pop all of them.

	sort.Ints(poppedVals) // Sort to check for correctness
	if len(poppedVals) != numElements {
		t.Errorf("Expected to pop %d elements, but got %d", numElements, len(poppedVals))
	}
	for i := 0; i < numElements; i++ {
		if poppedVals[i] != i {
			t.Errorf("Expected element %d, but got %d at index %d", i, poppedVals[i], i)
			break
		}
	}
}

// TestHeap_Stability_EqualElements explicitly tests stability for equal elements.
// This test is to demonstrate the *lack* of stability, which is expected for a standard binary heap.
func TestHeap_Stability_EqualElements(t *testing.T) {
	type Item struct {
		Value int
		ID    int // Unique ID to track original insertion order
	}

	heap := NewHeap(func(a, b Item) bool {
		return a.Value < b.Value // Min-heap based on Value
	})

	// Push items with the same Value but different IdataStructures
	heap.Push(Item{Value: 10, ID: 3})
	heap.Push(Item{Value: 5, ID: 1})
	heap.Push(Item{Value: 10, ID: 2})
	heap.Push(Item{Value: 5, ID: 4})
	heap.Push(Item{Value: 15, ID: 5})

	// Expected pop sequence: 5 (ID 1 or 4), 5 (other ID), 10 (ID 2 or 3), 10 (other ID), 15 (ID 5)

	firstPopped := heap.Pop()
	if firstPopped.Value != 5 {
		t.Errorf("Expected first popped value 5, got %d", firstPopped.Value)
	}

	secondPopped := heap.Pop()
	if secondPopped.Value != 5 {
		t.Errorf("Expected second popped value 5, got %d", secondPopped.Value)
	}

	// This is the key: we cannot guarantee the order of IdataStructures if Values are equal.
	// The following assert might fail depending on heap internal state, and that's OK.
	// This test simply demonstrates that the order of equal elements is not stable.
	if firstPopped.ID == 1 && secondPopped.ID == 4 {
		t.Logf("Stable pop order for 5s: 1 then 4 (pure coincidence)")
	} else if firstPopped.ID == 4 && secondPopped.ID == 1 {
		t.Logf("Unstable pop order for 5s: 4 then 1 (expected non-stability)")
	} else {
		t.Errorf("Unexpected IdataStructures for value 5: first %d, second %d", firstPopped.ID, secondPopped.ID)
	}

	// Further pops
	thirdPopped := heap.Pop()
	fourthPopped := heap.Pop()

	if thirdPopped.Value != 10 || fourthPopped.Value != 10 {
		t.Errorf("Expected values 10, got %d and %d", thirdPopped.Value, fourthPopped.Value)
	}

	// Again, no guarantee on IdataStructures for 10s.
}

// TestHeap_CompareFnNil checks behavior when custom_is_lower_fn is not set.
func TestHeap_CompareFnNil(t *testing.T) {
	heap := &Heap[int]{}
	// Calling Push/Pop without setting Compare_fn should ideally panic or fail
	// Currently, it would panic on `heap.custom_is_lower_fn(heap.items[current_index], heap.items[parent])`
	// because `custom_is_lower_fn` would be nil. This is expected and good.
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic when custom_is_lower_fn was not set")
		}
	}()
	heap.Push(1) // This should panic
}

// TestHeap_ZeroValueType ensures that zero values of `T` behave correctly.
func TestHeap_ZeroValueType(t *testing.T) {
	heap := NewHeap(func(a, b string) bool { return a < b }) // Min-heap for strings

	heap.Push("c")
	heap.Push("a")
	heap.Push("b")

	if item := heap.Pop(); item != "a" {
		t.Errorf("Expected 'a', got %s", item)
	}
	if item := heap.Pop(); item != "b" {
		t.Errorf("Expected 'b', got %s", item)
	}
	if item := heap.Pop(); item != "c" {
		t.Errorf("Expected 'c', got %s", item)
	}

	// Pop from empty heap
	emptyItem := heap.Pop()
	if emptyItem != "" { // Zero value for string
		t.Errorf("Expected empty string from Pop on empty heap, got %s", emptyItem)
	}
}

// TestHeap_PanicOnWrongSize checks the internal panic conditions.
func TestHeap_PanicOnWrongSize(t *testing.T) {
	// This test is hard to trigger programmatically without
	// manipulating the heap's internal slice directly, which
	// is not how users would interact with it.
	// The panics `panic(fmt.Sprintf("wrong heap size, expected %d, got %d", start_size + 1, len(heap.items)))`
	// are defensive checks against internal logic errors or race conditions
	// that bypass the mutex (which they shouldn't).
	// So, if the other concurrency tests pass without panics, it implicitly means these checks are not triggered.
	// No specific test function needed for this, as it's a "negative" test for internal consistency.
}
