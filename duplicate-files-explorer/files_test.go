package main

import (
	"archive-tools-monorepo/commons"
	"archive-tools-monorepo/commons/ds"
	"os"
	"path/filepath"
	"testing"
)

func TestCanFileBeRead(t *testing.T) {
	// Test Case 1: Valid File Path
	t.Run("ValidFilePath", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")

		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		defer os.Remove(tmpFile.Name())

		if err != nil {
			t.Fatalf("Failed to change file permissions: %v", err)
		}

		filePath := tmpFile.Name()

		if !can_file_be_read(&filePath) {
			t.Errorf("Expected %s to be a valid file path", filePath)
		}
	})

	// Test Case 2: Non-Existent File
	t.Run("NonExistentFile", func(t *testing.T) {
		nonExistentPath := "/nonexistent/file/path"
		if can_file_be_read(&nonExistentPath) {
			t.Errorf("Expected false for non-existent file, got true")
		}
	})

	// Test Case 3: File Without Read Permissions
	t.Run("FileWithoutReadPermissions", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		err = os.Chmod(tmpFile.Name(), 0200) // Write-only
		if err != nil {
			t.Fatalf("Failed to change file permissions: %v", err)
		}

		filePath := tmpFile.Name()
		if can_file_be_read(&filePath) {
			t.Errorf("Expected false for file without read permissions, got true")
		}
	})

	// Test Case 4: Directory Instead of File
	t.Run("DirectoryInsteadOfFile", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "testdir")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.Remove(tmpDir)
		dirPath := tmpDir
		if can_file_be_read(&dirPath) {
			t.Errorf("Expected false for directory, got true")
		}
	})

	// Test Case 5: Empty File
	t.Run("EmptyFile", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		filePath := tmpFile.Name()
		if !can_file_be_read(&filePath) {
			t.Errorf("Expected true for empty file, got false")
		}
	})

	// Test Case 6: File with Special Characters in Name
	t.Run("FileWithSpecialCharacters", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile-!@#$%^&*()")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())
		filePath := tmpFile.Name()
		if !can_file_be_read(&filePath) {
			t.Errorf("Expected true for file with special characters, got false")
		}
	})
}
func TestEvaluateObjectProperties(t *testing.T) {
	// Test Case 1: Symbolic Link
	t.Run("SymbolicLink", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		symlinkPath := tmpFile.Name() + "_symlink"
		err = os.Symlink(tmpFile.Name(), symlinkPath)
		if err != nil {
			t.Fatalf("Failed to create symbolic link: %v", err)
		}
		defer os.Remove(symlinkPath)

		info, err := os.Lstat(symlinkPath)
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}

		fullpath := symlinkPath
		result := evaluate_object_properties(&info, &fullpath)
		if result != symlink {
			t.Errorf("Expected invalid for symbolic link, got %d", result)
		}
	})

	// Test Case 2: Regular File
	t.Run("RegularFile", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		info, err := os.Stat(tmpFile.Name())
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}

		fullpath := tmpFile.Name()
		result := evaluate_object_properties(&info, &fullpath)
		if result != file {
			t.Errorf("Expected file for regular file, got %d", result)
		}
	})

	// Test Case 3: Directory
	t.Run("Directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "testdir")
		if err != nil {
			t.Fatalf("Failed to create temp directory: %v", err)
		}
		defer os.Remove(tmpDir)

		info, err := os.Stat(tmpDir)
		if err != nil {
			t.Fatalf("Failed to get directory info: %v", err)
		}

		fullpath := tmpDir
		result := evaluate_object_properties(&info, &fullpath)
		if result != directory {
			t.Errorf("Expected directory for directory, got %d", result)
		}
	})

	// Test Case 4: File Without Read Permissions
	t.Run("FileWithoutReadPermissions", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		err = os.Chmod(tmpFile.Name(), 0200) // Write-only
		if err != nil {
			t.Fatalf("Failed to change file permissions: %v", err)
		}

		info, err := os.Stat(tmpFile.Name())
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}

		fullpath := tmpFile.Name()
		result := evaluate_object_properties(&info, &fullpath)
		if result != invalid {
			t.Errorf("Expected invalid for file without read permissions, got %d", result)
		}
	})
}
func TestProcessFileEntry(t *testing.T) {
	t.Run("ValidFileEntry", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString("test content")
		if err != nil {
			t.Fatalf("Failed to write to temp file: %v", err)
		}

		info, err := os.Stat(tmpFile.Name())
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}

		fileHeap := &FileHeap{
			heap:           ds.Heap[commons.File]{},
			pending_insert: *ds.Build_new_atomic_counter(),
		}

		process_file_entry(filepath.Dir(tmpFile.Name()), &info, fileHeap)

		if ds.Is_heap_empty(&fileHeap.heap) {
			t.Errorf("Expected 1 file in heap, got 0",)
		}
	})

	t.Run("EmptyFileEntry", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		info, err := os.Stat(tmpFile.Name())
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}

		fileHeap := &FileHeap{
			heap:           ds.Heap[commons.File]{},
			pending_insert: *ds.Build_new_atomic_counter(),
		}

		process_file_entry(filepath.Dir(tmpFile.Name()), &info, fileHeap)

		if ds.Is_heap_empty(&fileHeap.heap) {
			t.Errorf("Expected 1 file in heap, got 0",)
		}
	})

	t.Run("UnreadableFileEntry", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "testfile")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		err = os.Chmod(tmpFile.Name(), 0200) // Write-only
		if err != nil {
			t.Fatalf("Failed to change file permissions: %v", err)
		}

		info, err := os.Stat(tmpFile.Name())
		if err != nil {
			t.Fatalf("Failed to get file info: %v", err)
		}

		fileHeap := &FileHeap{
			heap:           ds.Heap[commons.File]{},
			pending_insert: *ds.Build_new_atomic_counter(),
		}

		process_file_entry(filepath.Dir(tmpFile.Name()), &info, fileHeap)

		if !ds.Is_heap_empty(&fileHeap.heap) {
			t.Errorf("Expected 0 file in heap, got more",)
		}
	})
}

