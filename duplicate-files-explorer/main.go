package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	_ "embed"
	"flag"
	"os"
	"path/filepath"
	"time"
)

//go:embed semver.txt
var version string

//go:embed buildts.txt
var buildts string

var main_ui = commons.New_UI()

func main() {
	var basedir string
	var fullpath string
	var formatted_size commons.FileSize

	commons.Print_not_registered(main_ui, "Running version: %s", version)
	commons.Print_not_registered(main_ui, "Build timestamp: %s\n", buildts)

	commons.Register_new_line("directory-line", main_ui)
	commons.Register_new_line("file-line", main_ui)
	commons.Register_new_line("size-line", main_ui)

	flag.StringVar(&basedir, "dir", "", "Scan starting point  directory")
	flag.Parse()

	directories_stack := ds.Stack[string]{}
	output_file_heap := FileHeap{}

	ds.Set_compare_fn(&output_file_heap.heap, commons.Lower)
	output_file_heap.pending_insert = *ds.Build_new_atomic_counter()

	file_seen := 0
	directories_seen := 0
	size_processed := int64(0)
	cicles_counter := 0

	ds.Push_into_stack(&directories_stack, basedir)

	for !ds.Is_stack_empty(&directories_stack) {
		current_dir := ds.Pop_from_stack(&directories_stack)
		entries, read_dir_err := os.ReadDir(current_dir)

		if read_dir_err != nil {
			continue
		}

		for _, entry := range entries {
			fullpath = filepath.Join(current_dir, entry.Name())

			if entry.IsDir() {
				directories_seen += 1
				ds.Push_into_stack(&directories_stack, fullpath)
			} else {
				obj, err := entry.Info()

				if err == nil && evaluate_object_properties(&obj, &fullpath) == file {
					file_seen += 1
					size_processed += obj.Size()
					go process_file_entry(current_dir, &obj, &output_file_heap)

				}
			}
		}

		formatted_size = commons.Format_file_size(size_processed)

		cicles_counter++

		if cicles_counter > 50 {
			commons.Print_to_line(
				main_ui, "directory-line",
				"Directories seen: %6d", directories_seen,
			)

			commons.Print_to_line(
				main_ui, "file-line",
				"Files seen: %12d", file_seen,
			)

			commons.Print_to_line(
				main_ui, "size-line",
				"Processed: %10d %2s", formatted_size.Value,
				formatted_size.Unit,
			)

			cicles_counter = 0
		}
		apply_back_pressure(&output_file_heap.pending_insert)
	}

	commons.Print_to_line(
		main_ui, "directory-line",
		"Directories seen: %6d", directories_seen,
	)

	commons.Print_to_line(
		main_ui, "file-line",
		"Files seen: %12d", file_seen,
	)

	commons.Print_to_line(
		main_ui, "size-line",
		"Processed: %10d %2s", formatted_size.Value,
		formatted_size.Unit,
	)

	for ds.Get_counter_value(&output_file_heap.pending_insert) > 0 {
		time.Sleep(1 * time.Millisecond)
	}

	cleaned_heap_1 := build_duplicate_entries_heap(&output_file_heap.heap, true)
	cleaned_heap := build_duplicate_entries_heap(cleaned_heap_1, false)

	commons.Close_UI(main_ui)

	display_duplicate_file_info(cleaned_heap)
}
