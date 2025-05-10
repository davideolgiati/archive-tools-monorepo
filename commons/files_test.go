package commons

import "testing"

func TestCompareFiles(t *testing.T) {
	// Test Case 1: File A is less than File B
	t.Run("FileALessThanFileB", func(t *testing.T) {
		fileA := &File{Hash: "abc123", Size: 100}
		fileB := &File{Hash: "def456", Size: 200}

		if !Compare_files(fileA, fileB) {
			t.Errorf("Expected fileA to be less than fileB")
		}
	})

	// Test Case 2: File A is not less than File B
	t.Run("FileANotLessThanFileB", func(t *testing.T) {
		fileA := &File{Hash: "def456", Size: 200}
		fileB := &File{Hash: "abc123", Size: 100}

		if Compare_files(fileA, fileB) {
			t.Errorf("Expected fileA to not be less than fileB")
		}
	})

	// Test Case 3: File A and File B are equal
	t.Run("FileAEqualToFileB", func(t *testing.T) {
		fileA := &File{Hash: "abc123", Size: 100}
		fileB := &File{Hash: "abc123", Size: 100}

		if Compare_files(fileA, fileB) {
			t.Errorf("Expected fileA to not be less than fileB when they are equal")
		}
	})
}

func TestCheckIfFilesAreEqual(t *testing.T) {
	// Test Case 1: Files are equal
	t.Run("FilesAreEqual", func(t *testing.T) {
		fileA := &File{Hash: "abc123", Size: 100}
		fileB := &File{Hash: "abc123", Size: 100}

		if !Check_if_files_are_equal(fileA, fileB) {
			t.Errorf("Expected files to be equal")
		}
	})

	// Test Case 2: Files have different hashes
	t.Run("FilesHaveDifferentHashes", func(t *testing.T) {
		fileA := &File{Hash: "abc123", Size: 100}
		fileB := &File{Hash: "def456", Size: 100}

		if Check_if_files_are_equal(fileA, fileB) {
			t.Errorf("Expected files to not be equal due to different hashes")
		}
	})

	// Test Case 3: Files have different sizes
	t.Run("FilesHaveDifferentSizes", func(t *testing.T) {
		fileA := &File{Hash: "abc123", Size: 100}
		fileB := &File{Hash: "abc123", Size: 200}

		if Check_if_files_are_equal(fileA, fileB) {
			t.Errorf("Expected files to not be equal due to different sizes")
		}
	})

	// Test Case 4: Files have different hashes and sizes
	t.Run("FilesHaveDifferentHashesAndSizes", func(t *testing.T) {
		fileA := &File{Hash: "abc123", Size: 100}
		fileB := &File{Hash: "def456", Size: 200}

		if Check_if_files_are_equal(fileA, fileB) {
			t.Errorf("Expected files to not be equal due to different hashes and sizes")
		}
	})
}

func TestGetHumanReadableSizeAsync(t *testing.T) {
	// Test Case 1: Size in bytes
	t.Run("SizeInBytes", func(t *testing.T) {
		size := int64(500)
		expected := FileSize{Value: 500, Unit: "b"}
		result := Get_human_reabable_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test Case 2: Size in kilobytes
	t.Run("SizeInKilobytes", func(t *testing.T) {
		size := int64(1500)
		expected := FileSize{Value: 1, Unit: "Kb"}
		result := Get_human_reabable_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test Case 3: Size in megabytes
	t.Run("SizeInMegabytes", func(t *testing.T) {
		size := int64(2_500_000)
		expected := FileSize{Value: 2, Unit: "Mb"}
		result := Get_human_reabable_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test Case 4: Size in gigabytes
	t.Run("SizeInGigabytes", func(t *testing.T) {
		size := int64(5_000_000_000)
		expected := FileSize{Value: 5, Unit: "Gb"}
		result := Get_human_reabable_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test Case 5: Edge case - exactly 1000 bytes
	t.Run("EdgeCase1000Bytes", func(t *testing.T) {
		size := int64(1000)
		expected := FileSize{Value: 1, Unit: "Kb"}
		result := Get_human_reabable_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})

	// Test Case 6: Edge case - exactly 1,000,000 bytes
	t.Run("EdgeCase1000000Bytes", func(t *testing.T) {
		size := int64(1_000_000)
		expected := FileSize{Value: 1, Unit: "Mb"}
		result := Get_human_reabable_size(size)
		if result != expected {
			t.Errorf("Expected %v, got %v", expected, result)
		}
	})
}