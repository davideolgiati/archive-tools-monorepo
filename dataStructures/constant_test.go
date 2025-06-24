package datastructures_test

import (
	"fmt"
	"strings"
	"testing"

	datastructures "archive-tools-monorepo/dataStructures"
)

func TestConstant_NewConstant_WhenValueIsSet_ReturnOk(t *testing.T) {
	data := "test"
	constant, err := datastructures.NewConstant(&data)
	if err != nil {
		t.Errorf("%v", err)
	}

	if constant.Ptr() != &data {
		t.Errorf("Data pointer != Constant pointer (%x , %x)", &data, constant.Ptr())
	}

	if constant.Value() != data {
		t.Errorf("Data value != Constant value (%v, %v)", data, constant.Value())
	}
}

func TestConstant_NewConstant_WhenValueNotSet_Panic(t *testing.T) {
	var data *string
	_, err := datastructures.NewConstant(data)

	if err == nil {
		t.Error("Expected panic for nil pointer input, but got none")
	} else {
		if !strings.Contains(fmt.Sprintf("%v", err), "data pointer is nil") {
			t.Errorf("Unexpected panic message: %v", err)
		}
	}
}
