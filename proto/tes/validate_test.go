package tes

import "testing"

func TestValidation(t *testing.T) {
	v := Validate(&Task{})
	if len(v) == 0 {
		t.Fatal("expected validation errors")
	}
}
