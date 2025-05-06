package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	_ "embed"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

//go:embed semver.txt
var version string

//go:embed buildts.txt
var buildts string

var saveCursorPosition string = "\033[s"
var clearLine string = "\033[u\033[K"

func build_duplicate_entries_heap(file_heap *ds.Heap[commons.File]) *ds.Heap[commons.File] {
	var last_seen *commons.File
	var data *commons.File

	output := ds.Heap[commons.File]{}
	hash_channel := make(chan string)
	is_duplicate := false
	heap_size := ds.Get_heap_size(file_heap)
	ignored_files_counter := 0
	processed_files_counter := 0

	ds.Set_compare_fn(&output, commons.Compare_file_hashes)

	if !ds.Is_heap_empty(file_heap) {
		data = ds.Pop_from_heap(file_heap)
	}

	fmt.Print(saveCursorPosition)
	for !ds.Is_heap_empty(file_heap) {
		last_seen = data
		data = ds.Pop_from_heap(file_heap)
		processed_files_counter++

		if data.Hash == last_seen.Hash || is_duplicate {
			go commons.Hash_file(last_seen.Name, false, hash_channel)
			last_seen.Hash = <-hash_channel

			ds.Push_into_heap(&output, last_seen)
			is_duplicate = data.Hash == last_seen.Hash
		} else {
			ignored_files_counter++
		}

		fmt.Print(clearLine)
		fmt.Printf(
			"Removing unique entries ... %.2f %% (%6d files seen, %6d unique entries)",
			(float32(processed_files_counter)/float32(heap_size))*100, processed_files_counter,
			ignored_files_counter,
		)
	}

	fmt.Print("\n")

	return &output
}

func display_file_info_from_channel(
	file_heap *ds.Heap[commons.File],
) {
	var last_seen *commons.File
	var data *commons.File
	is_duplicate := false

	if !ds.Is_heap_empty(file_heap) {
		data = ds.Pop_from_heap(file_heap)
	}

	for !ds.Is_heap_empty(file_heap) {
		last_seen = data
		data = ds.Pop_from_heap(file_heap)

		if data.Hash == last_seen.Hash {
			print_file_details_to_stdout(last_seen)
			is_duplicate = true
		} else if is_duplicate {
			print_file_details_to_stdout(last_seen)
			is_duplicate = false
		}

	}

	if is_duplicate {
		print_file_details_to_stdout(last_seen)
	}
}

func compute_back_pressure(queue_size *int64) time.Duration {
	if *queue_size < 100 {
		return 0 * time.Millisecond
	}

	if *queue_size < 500 {
		return 1 * time.Millisecond
	}

	if *queue_size < 1000 {
		return 2 * time.Millisecond
	}

	return 3 * time.Millisecond
}

func check_if_file_is_valid(fullpath *string) bool {
	if commons.Is_file_symbolic_link(fullpath) {
		return false
	}

	if commons.Is_file_a_device(fullpath) {
		return false
	}

	if commons.Is_file_a_socket(fullpath) {
		return false
	}

	if commons.Is_file_a_pipe(fullpath) {
		return false
	}

	if !commons.Current_user_has_read_right_on_file(fullpath) {
		return false
	}

	return true
}

func main() {
	var basedir string
	var entry_name string
	var file_type fs.FileMode
	var fullpath string
	var formatted_size commons.FileSize
	var queue_size int64
	var back_pressure time.Duration

	main_ui := commons.New_UI()

	commons.Register_new_line("directory-line", main_ui)
	commons.Register_new_line("file-line", main_ui)
	commons.Register_new_line("size-line", main_ui)

	fmt.Printf("Running version: %s", version)
	fmt.Printf("Build timestamp: %s", buildts)

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	directories_stack := ds.Stack[string]{}
	output_file_heap := FileHeap{}

	ds.Set_compare_fn(&output_file_heap.heap, commons.Compare_file_hashes)
	output_file_heap.pending_insert = *ds.Create_new_atomic_counter()

	file_seen := 0
	directories_seen := 0
	size_processed := int64(0)

	ds.Push_into_stack(&directories_stack, basedir)

	fmt.Print(saveCursorPosition)
	for !ds.Is_stack_empty(&directories_stack) {
		current_dir := ds.Pop_from_stack(&directories_stack)
		entries, read_dir_err := os.ReadDir(current_dir)

		if read_dir_err != nil {
			continue
		}

		for _, entry := range entries {
			entry_info, file_info_err := entry.Info()

			if file_info_err != nil {
				continue
			}

			entry_name = entry_info.Name()
			file_type = entry_info.Mode()
			fullpath = filepath.Join(current_dir, entry_name)

			if !check_if_file_is_valid(&fullpath) {
				continue
			}

			if file_type.IsRegular() {
				file_seen += 1
				size_processed += entry_info.Size()
				go process_file_entry(&current_dir, &entry_info, &output_file_heap)
				commons.Print_to_line(
					main_ui, "file-line",
					"Files seen: %12d", file_seen,
				)
			}

			if file_type.IsDir() {
				directories_seen += 1
				ds.Push_into_stack(&directories_stack, fullpath)
				commons.Print_to_line(
					main_ui, "directory-line",
					"Directories seen: %6d", directories_seen,
				)
			}
		}

		formatted_size = commons.Get_human_reabable_size(size_processed)
		queue_size = ds.Get_counter_value(&output_file_heap.pending_insert)
		back_pressure = compute_back_pressure(&queue_size)

		commons.Print_to_line(
			main_ui, "size-line",
			"Processed: %10d %2s", formatted_size.Value,
			formatted_size.Unit,
		)
		
		time.Sleep(back_pressure)
	}

	commons.Close_UI(main_ui)

	for ds.Get_counter_value(&output_file_heap.pending_insert) > 0 {
		time.Sleep(1 * time.Millisecond)
	}

	cleaned_heap := build_duplicate_entries_heap(&output_file_heap.heap)
	display_file_info_from_channel(cleaned_heap)
}
