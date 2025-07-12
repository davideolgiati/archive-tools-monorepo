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

	myFile := commons.File{
		Name: "my/path/test",
		Hash: myHash,
		Size: 1000,
	}

	expected := "1234567890123                               1 Kb my/path/test"
	actualToString, err := myFile.ToString()
	actualFormat := fmt.Sprintf("%v", &myFile)

	if err != nil {
		panic(err)
	}

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
