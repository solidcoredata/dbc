package elist

import (
	"errors"
	"testing"
)

func TestEListWithError(t *testing.T) {
	var el EList
	el.Add(errors.New("error 1"))
	el.Add(errors.New("error 2"))
	if !el.IsError() {
		t.Fatal("error list should be an error")
	}
	if el.Error() != "error 1\nerror 2\n" {
		t.Fatalf("failed to list both errors: %s", el.Error())
	}
}

func TestEListNoError(t *testing.T) {
	el := EList{}
	if el.IsError() {
		t.Fatal("error list should not be error")
	}
}
