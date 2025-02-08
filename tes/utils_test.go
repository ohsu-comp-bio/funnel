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

	encoded, err := Base64Encode(task)
	if err != nil {
		t.Fatal(err)
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
