package tes

import (
	"reflect"
	"testing"
)

func TestBase64Encode(t *testing.T) {
	task := &Task{
		Id: "task1",
		Executors: []*Executor{
			{
				Image:   "alpine",
				Command: []string{"echo", "hello world"},
			},
		},
	}

	expected := "ewogICJpZCI6ICJ0YXNrMSIsCiAgImV4ZWN1dG9ycyI6IFsKICAgIHsKICAgICAgImltYWdlIjogImFscGluZSIsCiAgICAgICJjb21tYW5kIjogWwogICAgICAgICJlY2hvIiwKICAgICAgICAiaGVsbG8gd29ybGQiCiAgICAgIF0KICAgIH0KICBdCn0="

	encoded, err := Base64Encode(task)
	if err != nil {
		t.Fatal(err)
	}

	if encoded != expected {
		t.Logf("expected: %+v", expected)
		t.Logf("actual: %+v", encoded)
		t.Fatal("unexpected value returned from Base64Encode")
	}

	decoded, err := Base64Decode(encoded)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(decoded, task) {
		t.Logf("expected: %+v", task)
		t.Logf("actual: %+v", decoded)
		t.Fatal("incorrect decoded task from Base64Decode")
	}
}
