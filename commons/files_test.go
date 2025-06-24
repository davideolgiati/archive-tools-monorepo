package commons_test

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"archive-tools-monorepo/commons"
	datastructures "archive-tools-monorepo/dataStructures"
)

var sizesArray = [...]string{"b", "Kb", "Mb", "Gb"}

// TestFile_ToString verifies the formatting of the File struct.
func TestFile_ToString(t *testing.T) {
	hash := "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0"
	hashConstant, _ := datastructures.NewConstant(&hash)
	file := commons.File{
		Name: "test_document.txt",
		Hash: hashConstant,
		FormattedSize: commons.FileSize{
			Value: 123,
			Unit:  &sizesArray[1], // Kb
		},
		Size: 123456,
	}

	expected := "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0  123 Kb test_document.txt"
	if file.ToString() != expected {
		t.Errorf("ToString() mismatch.\nExpected: %q\nGot:      %q", expected, file.ToString())
	}

	// Test with different unit and hash length
	hash2 := "short"
	hashConstant, _ = datastructures.NewConstant(&hash2)
	file2 := commons.File{
		Name: "another_file.log",
		Hash: hashConstant,
		FormattedSize: commons.FileSize{
			Value: 5,
			Unit:  &sizesArray[0], // b
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
	f1 := commons.File{Name: "small.txt", Size: 100}
	f2 := commons.File{Name: "large.txt", Size: 200}

	if !commons.SizeDescending(f1, f2) { // 100 <= 200 -> true
		t.Errorf("Expected f1 to be <= f2")
	}
	if commons.SizeDescending(f2, f1) { // 200 <= 100 -> false
		t.Errorf("Expected f2 not to be <= f1")
	}
	if !commons.SizeDescending(f1, f1) { // 100 <= 100 -> true (crucial for equality)
		t.Errorf("Expected f1 to be <= f1 (equality check failed)")
	}

	// Test with equal sizes, different names
	f4 := commons.File{Name: "equal1.txt", Size: 100}
	f5 := commons.File{Name: "equal2.txt", Size: 100}

	// This function will return true for both SizeDescending(f4, f5) and SizeDescending(f5, f4)
	// if it is used for 'is_lower_fn' in a heap context, it would mean that f4 is not 'strictly' lower than f5,
	// and f5 is not 'strictly' lower than f4. This is where stability issues arise.
	if !commons.SizeDescending(f4, f5) {
		t.Errorf("Expected f4 to be <= f5 when sizes are equal")
	}
	if !commons.SizeDescending(f5, f4) {
		t.Errorf("Expected f5 to be <= f4 when sizes are equal")
	}
}

// TestHashDescending_Deterministic verifies comparison function for deterministic input.
func TestHashDescending_Deterministic(t *testing.T) {
	hash1 := "aaaa"
	hash2 := "bbbb"
	hashConstant1, _ := datastructures.NewConstant(&hash1)
	hashConstant2, _ := datastructures.NewConstant(&hash2)

	f1 := commons.File{Hash: hashConstant1}
	f2 := commons.File{Hash: hashConstant2}

	if !commons.HashDescending(&f1, &f2) { // "aaaa" <= "bbbb" -> true
		t.Errorf("Expected f1 to be <= f2")
	}
	if commons.HashDescending(&f2, &f1) { // "bbbb" <= "aaaa" -> false
		t.Errorf("Expected f2 not to be <= f1")
	}
	if !commons.HashDescending(&f1, &f1) { // "aaaa" <= "aaaa" -> true
		t.Errorf("Expected f1 to be <= f1 (equality check failed)")
	}

	// Test with equal hashes, different names (similar stability issue as SizeDescending)
	equalHash := "xyz"
	hashConstant3, _ := datastructures.NewConstant(&equalHash)
	hashConstant4, _ := datastructures.NewConstant(&equalHash)
	f4 := commons.File{Name: "f4", Hash: hashConstant3}
	f5 := commons.File{Name: "f5", Hash: hashConstant4}

	if !commons.HashDescending(&f4, &f5) {
		t.Errorf("Expected f4 to be <= f5 when hashes are equal")
	}
	if !commons.HashDescending(&f5, &f4) {
		t.Errorf("Expected f5 to be <= f4 when hashes are equal")
	}
}

// TestHash_Deterministic verifies hash generation is consistent for identical content.
func TestHash_Deterministic(t *testing.T) {
	// Create a temporary file with known content
	content := "This is some test content."
	tmpfile, err := os.CreateTemp(t.TempDir(), "test_hash_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove(tmpfile.Name())
		if err != nil {
			panic(err)
		}
	}()

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
	actualHash, err := commons.CalculateHash(tmpfile.Name())
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}

	if actualHash != expectedHash {
		t.Errorf("Hash mismatch.\nExpected: %s\nGot:      %s", expectedHash, actualHash)
	}

	// Test hashing an empty file
	emptyFile, err := os.CreateTemp(t.TempDir(), "test_empty_*.txt")
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		err := os.Remove(emptyFile.Name())
		if err != nil {
			panic(err)
		}
	}()

	err = emptyFile.Close()
	if err != nil {
		panic(err)
	}

	emptyHasher := sha1.New()
	expectedEmptyHash := hex.EncodeToString(emptyHasher.Sum(nil))

	actualEmptyHash, err := commons.CalculateHash(emptyFile.Name())
	if err != nil {
		t.Fatalf("Hash for empty file returned error: %v", err)
	}

	if actualEmptyHash != expectedEmptyHash {
		t.Errorf("Hash mismatch for empty file.\nExpected: %s\nGot:      %s", expectedEmptyHash, actualEmptyHash)
	}

	// Test non-existent file
	_, err = commons.CalculateHash("/path/to/non/existent/file.txt")

	if err == nil {
		t.Fatalf("Hash for unexistent file did not return error")
	}
}

// TestHash_Concurrent verifies Hash is safe for concurrent calls on DIFFERENT files.
func TestHash_Concurrent(t *testing.T) {
	numFiles := 100
	var filePaths []string
	var wg sync.WaitGroup

	// Create temporary files
	for i := range numFiles {
		content := fmt.Sprintf("content for file %d - %s", i, strings.Repeat("x", i))
		tmpfile, err := os.CreateTemp(t.TempDir(), fmt.Sprintf("concurrent_test_file_%d_*.txt", i))
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			err := os.Remove(tmpfile.Name())
			if err != nil {
				panic(err)
			}
		}()

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
			hash, err := commons.CalculateHash(p)
			if err != nil {
				panic(fmt.Sprintf("Hash for file returned error: %v", err))
			}

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
		contentBytes, err := os.ReadFile(path)
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
		expected commons.FileSize
	}{
		{0, commons.FileSize{Value: 0, Unit: &sizesArray[0]}},
		{500, commons.FileSize{Value: 500, Unit: &sizesArray[0]}},
		{1000, commons.FileSize{Value: 1, Unit: &sizesArray[1]}},
		{1023, commons.FileSize{Value: 1, Unit: &sizesArray[1]}}, // Still 1Kb
		{1234, commons.FileSize{Value: 1, Unit: &sizesArray[1]}},
		{999999, commons.FileSize{Value: 999, Unit: &sizesArray[1]}},
		{1000 * 1000, commons.FileSize{Value: 1, Unit: &sizesArray[2]}},
		{1500 * 1000, commons.FileSize{Value: 1, Unit: &sizesArray[2]}},
		{1000 * 1000 * 1000, commons.FileSize{Value: 1, Unit: &sizesArray[3]}},
		{5 * 1000 * 1000 * 1000, commons.FileSize{Value: 5, Unit: &sizesArray[3]}},
	}

	for _, tt := range tests {
		actual, err := commons.FormatFileSize(tt.input)
		if err != nil {
			t.Fatalf("Format file size returned error: %v", err)
		}

		if actual.Value != tt.expected.Value || *actual.Unit != *tt.expected.Unit {
			t.Errorf("Format_file_size(%d): Expected %v %s, Got %v %s",
				tt.input, tt.expected.Value, *tt.expected.Unit, actual.Value, *actual.Unit)
		}
	}

	_, err := commons.FormatFileSize(-100)

	if err == nil {
		t.Error("Expected panic for negative size input, but got none")
	} else {
		if !strings.Contains(fmt.Sprintf("%v", err), "size is negative") {
			t.Errorf("Unexpected panic message: %v", err)
		}
	}
}

// TestCheck_read_rights_on_file verifies file permission checking.
func TestCheck_read_rights_on_file(t *testing.T) {
	// Create a dummy file
	tmpfile, err := os.CreateTemp(t.TempDir(), "perms_test_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove(tmpfile.Name())
		if err != nil {
			panic(err)
		}
	}()

	// Get FileInfo
	info, err := os.Stat(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Test readable permissions (default for TempFile on Linux/macOS often includes read)
	if !commons.HasReadPermission(&info) {
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
	var nilInfo *os.FileInfo
	commons.HasReadPermission(nilInfo)
}

// TestIs_symbolic_link verifies symbolic link detection.
func TestIs_symbolic_link(t *testing.T) {
	// Create a regular file
	tmpfile, err := os.CreateTemp(t.TempDir(), "regular_file_*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := os.Remove(tmpfile.Name())
		if err != nil {
			panic(err)
		}
	}()

	err = tmpfile.Close()
	if err != nil {
		panic(err)
	}

	info, err := os.Stat(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}
	if commons.IsSymbolicLink(&info) {
		t.Errorf("Expected regular file not to be a symlink, but it is")
	}

	// Create a symbolic link (skip if not on Unix-like or permissions issue)
	symlinkPath := tmpfile.Name() + ".symlink"
	err = os.Symlink(tmpfile.Name(), symlinkPath)
	if err != nil {
		t.Skipf("Could not create symlink (e.g., Windows without admin, or permissions): %v", err)
	}
	defer func() {
		err := os.Remove(symlinkPath) // Clean up symlink
		if err != nil {
			panic(err)
		}
	}()

	symlinkInfo, err := os.Lstat(symlinkPath) // Use Lstat for symlink info
	if err != nil {
		t.Fatal(err)
	}

	if !commons.IsSymbolicLink(&symlinkInfo) {
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
	var nilInfo *os.FileInfo
	commons.IsSymbolicLink(nilInfo)
}
