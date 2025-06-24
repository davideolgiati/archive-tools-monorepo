package datastructures_test

import (
	"testing"

	datastructures "archive-tools-monorepo/dataStructures"
)

func TestFlyweight_SameValue_ExpectToBeEqual(t *testing.T) {
	fw := datastructures.Flyweight[string]{}
	constant1, err := fw.Instance("test")
	if err != nil {
		t.Fatalf("Instance returned error: %v", err)
	}

	constant2, err := fw.Instance("test")
	if err != nil {
		t.Fatalf("Instance returned error: %v", err)
	}

	if constant1 != constant2 {
		t.Errorf("Expected the same constant for identical inputs, got %v and %v", constant1, constant2)
	}
}

func TestFlyweight_DifferentValue_ExpectToBeDifferent(t *testing.T) {
	fw := datastructures.Flyweight[string]{}
	constant1, err := fw.Instance("test")
	if err != nil {
		t.Fatalf("Instance returned error: %v", err)
	}

	constant2, err := fw.Instance("test2")
	if err != nil {
		t.Fatalf("Instance returned error: %v", err)
	}

	if constant1 == constant2 {
		t.Errorf("Expected different constants for different inputs, got %v and %v", constant1, constant2)
	}
}
