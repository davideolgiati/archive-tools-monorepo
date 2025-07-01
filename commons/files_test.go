package commons_test

import (
	"fmt"
	"strings"
	"testing"

	"archive-tools-monorepo/commons"
	datastructures "archive-tools-monorepo/dataStructures"
)

func TestFile_NewFile_InstanceAndFormat_Ok(t *testing.T) {
	myHashValue := "1234567890123"
	myHash, err := datastructures.NewConstant(&myHashValue)
	if err != nil {
		panic(err)
	}

	myFileSize, err := commons.FormatFileSize(1000)
	if err != nil {
		panic(err)
	}

	myFile := commons.File{
		Name:          "my/path/test",
		Hash:          myHash,
		FormattedSize: myFileSize,
		Size:          1000,
	}

	expected := "1234567890123                               1 Kb my/path/test"
	actualToString := myFile.ToString()
	actualFormat := fmt.Sprintf("%v", myFile)

	if expected != actualToString {
		t.Errorf("expecting \"%v\", got \"%v\"", expected, actualToString)
	}

	if expected != actualFormat {
		t.Errorf("expecting \"%v\", got \"%v\"", expected, actualFormat)
	}
}

func TestFile_FormatFileSize_NegativeSize_Error(t *testing.T) {
	expected := commons.FileSize{}
	myFileSize, err := commons.FormatFileSize(-1000)

	if myFileSize != expected {
		t.Errorf("expected empty formatted FileSize, got %v", myFileSize)
	}

	if err == nil {
		t.Error("expected error to be valorized, got nil")
	}

	if !strings.Contains(fmt.Sprintf("%v", err), "invalid argument: size is negative") {
		t.Errorf("expected error to be \"invalid argument: size is negative\", got %v", err)
	}
}

func TestFile_SizeDescending_Ok(t *testing.T) {
	myHashValue := "1234567890123"
	myHash, err := datastructures.NewConstant(&myHashValue)
	if err != nil {
		panic(err)
	}

	myFileSize1, err := commons.FormatFileSize(1000)
	if err != nil {
		panic(err)
	}

	myFileSize2, err := commons.FormatFileSize(2000)
	if err != nil {
		panic(err)
	}

	myFile1 := commons.File{
		Name:          "my/path/test",
		Hash:          myHash,
		FormattedSize: myFileSize1,
		Size:          1000,
	}

	myFile2 := commons.File{
		Name:          "my/path/test",
		Hash:          myHash,
		FormattedSize: myFileSize2,
		Size:          2000,
	}

	myFile3 := commons.File{
		Name:          "my/path/test",
		Hash:          myHash,
		FormattedSize: myFileSize2,
		Size:          -2000,
	}

	compare1, err := commons.SizeDescending(myFile2, myFile1)
	if err != nil {
		t.Errorf("commons.SizeDescending(myFile2, myFile1) expected to return nil err, got  %v", err)
	}

	compare2, err := commons.SizeDescending(myFile1, myFile2)
	if err != nil {
		t.Errorf("commons.SizeDescending(myFile1, myFile2) expected to return nil err, got  %v", err)
	}

	if compare1 {
		t.Errorf("Expected file2 (2000) to be > file1 (1000)")
	}

	if !compare2 {
		t.Errorf("Expected file2 (1000) to be < file1 (2000)")
	}

	_, err = commons.SizeDescending(myFile1, myFile3)

	if err == nil {
		t.Error("commons.SizeDescending(myFile1, myFile3) expected to return error, got  nil")
	}

	if !strings.Contains(fmt.Sprintf("%v", err), "invalid argument: b.Size is negative") {
		t.Errorf("expected error to be \"invalid argument: b.Size is negative\", got %v", err)
	}

	_, err = commons.SizeDescending(myFile3, myFile1)

	if err == nil {
		t.Error("commons.SizeDescending(myFile3, myFile1) expected to return error, got  nil")
	}

	if !strings.Contains(fmt.Sprintf("%v", err), "invalid argument: a.Size is negative") {
		t.Errorf("expected error to be \"invalid argument: a.Size is negative\", got %v", err)
	}

	compare3, err := commons.SizeDescending(myFile1, myFile1)
	if err != nil {
		t.Errorf("commons.SizeDescending(myFile1, myFile1) expected to return nil err, got  %v", err)
	}

	if !compare3 {
		t.Errorf("Expected file1 (2000) to be < file1 (2000)")
	}
}
