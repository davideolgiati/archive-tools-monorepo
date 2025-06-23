package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDirWalker_FileAndDirectoryFilters(t *testing.T) {
	// Create a temporary directory structure.
	baseDir, err := os.MkdirTemp("", "dirwalker_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := os.RemoveAll(baseDir)

		if err != nil {
			panic(err)
		}
	}()

	// Create a subdirectory.
	subDir := filepath.Join(baseDir, "sub")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Create files in baseDir and subDir.
	file1 := filepath.Join(baseDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")
	emptyFile := filepath.Join(subDir, "empty.txt")
	if err := os.WriteFile(file1, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(file2, []byte("world"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a new walker with skip_empty=true.
	walker := NewWalker(true)
	walker.SetEntryPoint(baseDir)
	// Directory filter: allow all directories.
	walker.SetDirectoryFilter(func(dir string) bool {
		return true
	})

	processedFiles := []File{}
	walker.SetFileCallback(func(info File) {
		processedFiles = append(processedFiles, info)
	})

	walker.SetDirectoryCallback(func() {})

	// Execute the walk.
	walker.Walk()

	// Expect file1.txt and file2.txt to be processed (empty file skipped).
	expected := map[string]bool{
		file1: true,
		file2: true,
	}

	if len(processedFiles) != 2 {
		t.Errorf("expected 2 files processed, got %d", len(processedFiles))
	}
	for _, file := range processedFiles {
		if !expected[file.path] {
			t.Errorf("unexpected processed file: %s", file)
		}
	}
}

func TestDirWalker_SkipEmptyFiles(t *testing.T) {
	// Temporary directory with one non-empty and one empty file.
	baseDir, err := os.MkdirTemp("", "dirwalker_skipempty_test")

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := os.RemoveAll(baseDir)

		if err != nil {
			panic(err)
		}
	}()

	nonEmptyFile := filepath.Join(baseDir, "nonempty.txt")
	emptyFile := filepath.Join(baseDir, "empty.txt")
	if err := os.WriteFile(nonEmptyFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	walker := NewWalker(true)
	walker.SetEntryPoint(baseDir)
	walker.SetDirectoryFilter(func(dir string) bool {
		return true
	})

	processedFiles := []File{}
	walker.SetFileCallback(func(info File) {
		processedFiles = append(processedFiles, info)
	})

	walker.SetDirectoryCallback(func() {})

	walker.Walk()

	// Should process only the non-empty file.
	if len(processedFiles) != 1 {
		t.Errorf("expected 1 file processed, got %d", len(processedFiles))
	}
	if processedFiles[0].path != nonEmptyFile {
		t.Errorf("expected non-empty file to be processed")
	}
}

func TestDirWalker_NoSkipEmptyFiles(t *testing.T) {
	// Temporary directory with one non-empty and one empty file.
	baseDir, err := os.MkdirTemp("", "dirwalker_noskipempty_test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err = os.RemoveAll(baseDir)

		if err != nil {
			panic(err)
		}
	}()

	nonEmptyFile := filepath.Join(baseDir, "nonempty.txt")
	emptyFile := filepath.Join(baseDir, "empty.txt")
	if err := os.WriteFile(nonEmptyFile, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(emptyFile, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	walker := NewWalker(false) // Do not skip empty files.
	walker.SetEntryPoint(baseDir)
	walker.SetDirectoryFilter(func(dir string) bool {
		return true
	})

	processedFiles := []File{}
	walker.SetFileCallback(func(info File) {
		processedFiles = append(processedFiles, info)
	})

	walker.SetDirectoryCallback(func() {})

	walker.Walk()

	// Should process both files.
	if len(processedFiles) != 2 {
		t.Errorf("expected 2 files processed, got %d", len(processedFiles))
	}
}
