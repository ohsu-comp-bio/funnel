package tes

import "testing"

func TestValidation(t *testing.T) {
	v := Validate(&Task{})
	if len(v) == 0 {
		t.Fatal("expected validation errors")
	}
}

func TestEmptyTagKeyValidation(t *testing.T) {
	v := Validate(&Task{
		Tags: map[string]string{
			"": "bar",
		},
		Executors: []*Executor{
			{
				Image:   "alpine",
				Command: []string{"echo"},
			},
		},
	})
	if len(v) != 1 {
		t.Fatal("expected 1 validation error")
	}
}
