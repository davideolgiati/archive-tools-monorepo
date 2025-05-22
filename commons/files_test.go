package commons

import (
	"crypto/sha1"
	"fmt"
	"hash/crc32"
	"os"
	"testing"
)

func strPtr(s string) *string {
	return &s
}

func TestCompareFiles(t *testing.T) {
	// Test Case 1: File A is less than File B
	t.Run("FileALessThanFileB", func(t *testing.T) {
		fileA := File{Hash: strPtr("abc123"), Size: 100}
		fileB := File{Hash: strPtr("def456"), Size: 200}

		if !Lower(fileA, fileB) {
			t.Errorf("Expected fileA to be less than fileB")
		}
	})

	// Test Case 2: File A is not less than File B
	t.Run("FileANotLessThanFileB", func(t *testing.T) {
		fileA := File{Hash: strPtr("def456"), Size: 200}
		fileB := File{Hash: strPtr("abc123"), Size: 100}

		if Lower(fileA, fileB) {
			t.Errorf("Expected fileA to not be less than fileB")
		}
	})

	// Test Case 3: File A and File B are equal
	t.Run("FileAEqualToFileB", func(t *testing.T) {
		fileA := File{Hash: strPtr("abc123"), Size: 100}
		fileB := File{Hash: strPtr("abc123"), Size: 100}

		if Lower(fileA, fileB) {
			t.Errorf("Expected fileA to not be less than fileB when they are equal")
		}
	})
}

func TestCheckIfFilesAreEqual(t *testing.T) {
	// Test Case 1: Files are equal
	t.Run("FilesAreEqual", func(t *testing.T) {
		fileA := File{Hash: strPtr("abc123"), Size: 100}
		fileB := File{Hash: strPtr("abc123"), Size: 100}

		if !Equal(fileA, fileB) {
			t.Errorf("Expected files to be equal")
		}
	})

	// Test Case 2: Files have different hashes
	t.Run("FilesHaveDifferentHashes", func(t *testing.T) {
		fileA := File{Hash: strPtr("abc123"), Size: 100}
		fileB := File{Hash: strPtr("def456"), Size: 100}

		if Equal(fileA, fileB) {
			t.Errorf("Expected files to not be equal due to different hashes")
		}
	})

	// Test Case 3: Files have different sizes
	t.Run("FilesHaveDifferentSizes", func(t *testing.T) {
		fileA := File{Hash: strPtr("abc123"), Size: 100}
		fileB := File{Hash: strPtr("abc123"), Size: 200}

		if Equal(fileA, fileB) {
			t.Errorf("Expected files to not be equal due to different sizes")
		}
	})

	// Test Case 4: Files have different hashes and sizes
	t.Run("FilesHaveDifferentHashesAndSizes", func(t *testing.T) {
		fileA := File{Hash: strPtr("abc123"), Size: 100}
		fileB := File{Hash: strPtr("def456"), Size: 200}

		if Equal(fileA, fileB) {
			t.Errorf("Expected files to not be equal due to different hashes and sizes")
		}
	})

	// Test Case 1: Size in bytes
	t.Run("SizeInBytes", func(t *testing.T) {
		size := int64(500)
		expected := FileSize{Value: 500, Unit: strPtr("b")}
		result := Format_file_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test Case 2: Size in kilobytes
	t.Run("SizeInKilobytes", func(t *testing.T) {
		size := int64(1500)
		expected := FileSize{Value: 1, Unit: strPtr("Kb")}
		result := Format_file_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test Case 3: Size in megabytes
	t.Run("SizeInMegabytes", func(t *testing.T) {
		size := int64(2_500_000)
		expected := FileSize{Value: 2, Unit: strPtr("Mb")}
		result := Format_file_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	// Test Case 4: Size in gigabytes
	t.Run("SizeInGigabytes", func(t *testing.T) {
		size := int64(5_000_000_000)
		expected := FileSize{Value: 5, Unit: strPtr("Gb")}
		result := Format_file_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test Case 5: Edge case - exactly 1000 bytes
	t.Run("EdgeCase1000Bytes", func(t *testing.T) {
		size := int64(1000)
		expected := FileSize{Value: 1, Unit: strPtr("Kb")}
		result := Format_file_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
	
	// Test Case 6: Edge case - exactly 1,000,000 bytes
	t.Run("EdgeCase1000000Bytes", func(t *testing.T) {
		size := int64(1_000_000)
		expected := FileSize{Value: 1, Unit: strPtr("Mb")}
		result := Format_file_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}
func TestHash(t *testing.T) {
	// Test Case 1: Hash a small file with quick_flag set to true
	t.Run("HashSmallFileQuickFlag", func(t *testing.T) {
		filepath := "testfile_small.txt"
		content := "This is a test file."
		os.WriteFile(filepath, []byte(content), 0644)
		defer os.Remove(filepath)

		hash, err := Hash(filepath, int64(len(content)))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedHash := fmt.Sprintf("%x", crc32.ChecksumIEEE([]byte(content[:page_size*5])))
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})

	// Test Case 2: Hash a small file with quick_flag set to false
	t.Run("HashSmallFileFullHash", func(t *testing.T) {
		filepath := "testfile_small.txt"
		content := "This is a test file."
		os.WriteFile(filepath, []byte(content), 0644)
		defer os.Remove(filepath)

		hash, err := Hash(filepath, int64(len(content)))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedHash := fmt.Sprintf("%x", sha1.Sum([]byte(content)))
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})

	// Test Case 3: Hash a large file with quick_flag set to true
	t.Run("HashLargeFileQuickFlag", func(t *testing.T) {
		filepath := "testfile_large.txt"
		content := make([]byte, page_size*10)
		for i := range content {
			content[i] = byte(i % 256)
		}
		os.WriteFile(filepath, content, 0644)
		defer os.Remove(filepath)

		hash, err := Hash(filepath, int64(len(content)))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedHash := fmt.Sprintf("%x", crc32.ChecksumIEEE(content[:page_size*5]))
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})

	// Test Case 4: Hash a large file with quick_flag set to false
	t.Run("HashLargeFileFullHash", func(t *testing.T) {
		filepath := "testfile_large.txt"
		content := make([]byte, page_size*10)
		for i := range content {
			content[i] = byte(i % 256)
		}
		os.WriteFile(filepath, content, 0644)
		defer os.Remove(filepath)

		hash, err := Hash(filepath, int64(len(content)))
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		expectedHash := fmt.Sprintf("%x", sha1.Sum(content))
		if hash != expectedHash {
			t.Errorf("Expected hash %s, got %s", expectedHash, hash)
		}
	})

	// Test Case 5: File does not exist
	t.Run("FileDoesNotExist", func(t *testing.T) {
		filepath := "nonexistent_file.txt"
		_, err := Hash(filepath, 0)
		if err == nil {
			t.Errorf("Expected an error for nonexistent file, got nil")
		}
	})
}

