package commons

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
)
/*
// TestFile_ToString verifies the formatting of the File struct.
func TestFile_ToString(t *testing.T) {
	hash := "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0"
	file := File{
		Name: "test_document.txt",
		Hash: &hash,
		FormattedSize: FileSize{
			Value: 123,
			Unit:  &sizes_array[1], // Kb
		},
		Size: 123456,
	}

	expected := "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0  123 Kb test_document.txt"
	if file.ToString() != expected {
		t.Errorf("ToString() mismatch.\nExpected: %q\nGot:      %q", expected, file.ToString())
	}

	// Test with different unit and hash length
	hash2 := "short"
	file2 := File{
		Name: "another_file.log",
		Hash: &hash2,
		FormattedSize: FileSize{
			Value: 5,
			Unit:  &sizes_array[0], // b
		},
		Size: 50,
	}
	expected2 := "short                                       5  b another_file.log" // Check padding
	if file2.ToString() != expected2 {
		t.Errorf("ToString() mismatch with short hash/unit.\nExpected: %q\nGot:      %q", expected2, file2.ToString())
	}
}

// TestSizeDescending_Deterministic verifies comparison function for deterministic input.
func TestSizeDescending_Deterministic(t *testing.T) {
	f1 := File{Name: "small.txt", Size: 100}
	f2 := File{Name: "large.txt", Size: 200}

	if !SizeDescending(f1, f2) { // 100 <= 200 -> true
		t.Errorf("Expected f1 to be <= f2")
	}
	if SizeDescending(f2, f1) { // 200 <= 100 -> false
		t.Errorf("Expected f2 not to be <= f1")
	}
	if !SizeDescending(f1, f1) { // 100 <= 100 -> true (crucial for equality)
		t.Errorf("Expected f1 to be <= f1 (equality check failed)")
	}

	// Test with equal sizes, different names
	f4 := File{Name: "equal1.txt", Size: 100}
	f5 := File{Name: "equal2.txt", Size: 100}

	// This function will return true for both SizeDescending(f4, f5) and SizeDescending(f5, f4)
	// if it is used for 'is_lower_fn' in a heap context, it would mean that f4 is not 'strictly' lower than f5,
	// and f5 is not 'strictly' lower than f4. This is where stability issues arise.
	if !SizeDescending(f4, f5) {
		t.Errorf("Expected f4 to be <= f5 when sizes are equal")
	}
	if !SizeDescending(f5, f4) {
		t.Errorf("Expected f5 to be <= f4 when sizes are equal")
	}
}

// TestHashDescending_Deterministic verifies comparison function for deterministic input.
func TestHashDescending_Deterministic(t *testing.T) {
	hash1 := "aaaa"
	hash2 := "bbbb"

	f1 := File{Hash: &hash1}
	f2 := File{Hash: &hash2}

	if !HashDescending(f1, f2) { // "aaaa" <= "bbbb" -> true
		t.Errorf("Expected f1 to be <= f2")
	}
	if HashDescending(f2, f1) { // "bbbb" <= "aaaa" -> false
		t.Errorf("Expected f2 not to be <= f1")
	}
	if !HashDescending(f1, f1) { // "aaaa" <= "aaaa" -> true
		t.Errorf("Expected f1 to be <= f1 (equality check failed)")
	}

	// Test with equal hashes, different names (similar stability issue as SizeDescending)
	equalHash := "xyz"
	f4 := File{Name: "f4", Hash: ds.Constant[string]{&equalHash}}
	f5 := File{Name: "f5", Hash: ds.Constant[string]{ptr: &equalHash}}

	if !HashDescending(f4, f5) {
		t.Errorf("Expected f4 to be <= f5 when hashes are equal")
	}
	if !HashDescending(f5, f4) {
		t.Errorf("Expected f5 to be <= f4 when hashes are equal")
	}
}
*/

// TestHash_Deterministic verifies hash generation is consistent for identical content.
func TestHash_Deterministic(t *testing.T) {
	// Create a temporary file with known content
	content := "This is some test content."
	tmpfile, err := ioutil.TempFile("", "test_hash_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name()) // Clean up

	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Calculate expected SHA-1 hash manually
	hasher := sha1.New()
	hasher.Write([]byte(content))
	expectedHash := hex.EncodeToString(hasher.Sum(nil))

	// Get hash using the function
	actualHash := Hash(tmpfile.Name(), int64(len(content)))
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}

	if actualHash != expectedHash {
		t.Errorf("Hash mismatch.\nExpected: %s\nGot:      %s", expectedHash, actualHash)
	}

	// Test hashing an empty file
	emptyFile, err := ioutil.TempFile("", "test_empty_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(emptyFile.Name())
	emptyFile.Close()

	emptyHasher := sha1.New()
	expectedEmptyHash := hex.EncodeToString(emptyHasher.Sum(nil))

	actualEmptyHash := Hash(emptyFile.Name(), 0)
	if err != nil {
		t.Fatalf("Hash for empty file returned error: %v", err)
	}
	if actualEmptyHash != expectedEmptyHash {
		t.Errorf("Hash mismatch for empty file.\nExpected: %s\nGot:      %s", expectedEmptyHash, actualEmptyHash)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic when path to non existent file was given")
		}
	}()

	// Test non-existent file
	_ = Hash("/path/to/non/existent/file.txt", 0)
}

// TestHash_Concurrent verifies Hash is safe for concurrent calls on DIFFERENT files.
func TestHash_Concurrent(t *testing.T) {
	numFiles := 100
	var filePaths []string
	var wg sync.WaitGroup

	// Create temporary files
	for i := 0; i < numFiles; i++ {
		content := fmt.Sprintf("content for file %d - %s", i, strings.Repeat("x", i))
		tmpfile, err := ioutil.TempFile("", fmt.Sprintf("concurrent_test_file_%d_*.txt", i))
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())
		if _, err := tmpfile.WriteString(content); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}
		filePaths = append(filePaths, tmpfile.Name())
	}

	results := make(chan struct {
		path string
		hash string
	}, numFiles)

	// Concurrently hash files
	for _, path := range filePaths {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			stats, _ := os.Stat(p)
			hash := Hash(p, stats.Size())
			results <- struct {
				path string
				hash string
			}{p, hash}
		}(path)
	}

	wg.Wait()
	close(results)

	hashedFiles := make(map[string]string)
	for res := range results {
		hashedFiles[res.path] = res.hash
	}

	// Verify hashes by re-calculating them sequentially
	for _, path := range filePaths {
		contentBytes, err := ioutil.ReadFile(path)
		if err != nil {
			t.Errorf("Failed to read file %s for verification: %v", path, err)
			continue
		}
		hasher := sha1.New()
		hasher.Write(contentBytes)
		expectedHash := hex.EncodeToString(hasher.Sum(nil))

		if hashedFiles[path] != expectedHash {
			t.Errorf("Concurrent hash mismatch for file %s.\nExpected: %s\nGot:      %s", path, expectedHash, hashedFiles[path])
		}
	}
}

// TestFormat_file_size_Deterministic verifies size formatting.
func TestFormat_file_size_Deterministic(t *testing.T) {
	tests := []struct {
		input    int64
		expected FileSize
	}{
		{0, FileSize{Value: 0, Unit: &sizes_array[0]}},
		{500, FileSize{Value: 500, Unit: &sizes_array[0]}},
		{1000, FileSize{Value: 1, Unit: &sizes_array[1]}},
		{1023, FileSize{Value: 1, Unit: &sizes_array[1]}}, // Still 1Kb
		{1234, FileSize{Value: 1, Unit: &sizes_array[1]}},
		{999999, FileSize{Value: 999, Unit: &sizes_array[1]}},
		{1000 * 1000, FileSize{Value: 1, Unit: &sizes_array[2]}},
		{1500 * 1000, FileSize{Value: 1, Unit: &sizes_array[2]}},
		{1000 * 1000 * 1000, FileSize{Value: 1, Unit: &sizes_array[3]}},
		{5 * 1000 * 1000 * 1000, FileSize{Value: 5, Unit: &sizes_array[3]}},
	}

	for _, tt := range tests {
		actual := Format_file_size(tt.input)
		if actual.Value != tt.expected.Value || *actual.Unit != *tt.expected.Unit {
			t.Errorf("Format_file_size(%d): Expected %v %s, Got %v %s",
				tt.input, tt.expected.Value, *tt.expected.Unit, actual.Value, *actual.Unit)
		}
	}

	// Test negative size (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for negative size input, but got none")
		} else {
			if !strings.Contains(fmt.Sprintf("%v", r), "size is negative") {
				t.Errorf("Unexpected panic message: %v", r)
			}
		}
	}()
	Format_file_size(-100)
}

// TestCheck_read_rights_on_file verifies file permission checking.
func TestCheck_read_rights_on_file(t *testing.T) {
	// Create a dummy file
	tmpfile, err := ioutil.TempFile("", "perms_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	// Get FileInfo
	info, err := os.Stat(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Test readable permissions (default for TempFile on Linux/macOS often includes read)
	if !Check_read_rights_on_file(&info) {
		t.Errorf("Expected file to be readable by default, got false")
	}

	// Test with nil FileInfo (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil obj input, but got none")
		} else {
			if !strings.Contains(fmt.Sprintf("%v", r), "obj is nil") {
				t.Errorf("Unexpected panic message: %v", r)
			}
		}
	}()
	var nilInfo *os.FileInfo = nil
	Check_read_rights_on_file(nilInfo)
}

// TestIs_symbolic_link verifies symbolic link detection.
func TestIs_symbolic_link(t *testing.T) {
	// Create a regular file
	tmpfile, err := ioutil.TempFile("", "regular_file_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	info, err := os.Stat(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if Is_symbolic_link(&info) {
		t.Errorf("Expected regular file not to be a symlink, but it is")
	}

	// Create a symbolic link (skip if not on Unix-like or permissions issue)
	symlinkPath := tmpfile.Name() + ".symlink"
	err = os.Symlink(tmpfile.Name(), symlinkPath)
	if err != nil {
		t.Skipf("Could not create symlink (e.g., Windows without admin, or permissions): %v", err)
	}
	defer os.Remove(symlinkPath) // Clean up symlink

	symlinkInfo, err := os.Lstat(symlinkPath) // Use Lstat for symlink info
	if err != nil {
		t.Fatal(err)
	}

	if !Is_symbolic_link(&symlinkInfo) {
		t.Errorf("Expected symlink to be detected as symlink, but it is not")
	}

	// Test with nil FileInfo (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nil obj input, but got none")
		} else {
			if !strings.Contains(fmt.Sprintf("%v", r), "obj is nil") {
				t.Errorf("Unexpected panic message: %v", r)
			}
		}
	}()
	var nilInfo *os.FileInfo = nil
	Is_symbolic_link(nilInfo)
}
