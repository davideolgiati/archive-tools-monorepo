package datastructures

import (
	datastructures "archive-tools-monorepo/dataStructures"
	"fmt"
	"strings"
	"testing"
)

func TestConstant_NewConstant_WhenValueIsSet_ReturnOk(t *testing.T) {
	data := "test"
	constant, err := datastructures.NewConstant(&data)

	if err != nil {
		t.Errorf("%v", err)
	}

	if constant.Ptr() != &data {
		t.Errorf("Data pointer != Constant pointer (%x , %x)", constant.Ptr(), &data)
	}
}

func TestConstant_NewConstant_WhenValueNotSet_Panic(t *testing.T) {
	var data *string = nil
	_, err := datastructures.NewConstant(data)

	if err == nil {
		t.Error("Expected panic for nil pointer input, but got none")
	} else {
		if !strings.Contains(fmt.Sprintf("%v", err), "data pointer is nil") {
			t.Errorf("Unexpected panic message: %v", err)
		}
	}
}
