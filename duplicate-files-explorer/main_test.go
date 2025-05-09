package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"testing"
	"time"
)

func TestBuildDuplicateEntriesHeap(t *testing.T) {
	fileHeap := ds.Heap[commons.File]{}
	ds.Set_compare_fn(&fileHeap, commons.Compare_file_hashes)

	// Add test data
	ds.Push_into_heap(&fileHeap, &commons.File{Name: "../tmp-test-dir/heap_test/file1", Hash: "hash1"})
	ds.Push_into_heap(&fileHeap, &commons.File{Name: "../tmp-test-dir/heap_test/file2", Hash: "hash1"})
	ds.Push_into_heap(&fileHeap, &commons.File{Name: "../tmp-test-dir/heap_test/file3", Hash: "hash2"})

	result := build_duplicate_entries_heap(&fileHeap)

	if ds.Get_heap_size(result) != 2 {
		t.Errorf("Expected 2 duplicate entries, got %d", ds.Get_heap_size(result))
	}
}

func TestComputeBackPressure(t *testing.T) {
	tests := []struct {
		queueSize int64
		expected  time.Duration
	}{
		{50, 0 * time.Millisecond},
		{200, 1 * time.Millisecond},
		{700, 2 * time.Millisecond},
		{1500, 3 * time.Millisecond},
	}

	for _, test := range tests {
		counter := ds.AtomicCounter{}
		
		for i := int64(0); i < test.queueSize; i++ {
			ds.Increment(&counter)
		}

		result := compute_back_pressure(&counter)
		
		if result != test.expected {
			t.Errorf("For queue size %d, expected %v, got %v", test.queueSize, test.expected, result)
		}
	}
}
