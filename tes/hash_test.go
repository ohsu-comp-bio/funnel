package tes

import (
	"testing"
)

func TestHash(t *testing.T) {
	tests := []struct {
		task *Task
		hash string
	}{
		{
			&Task{
				Executors: []*Executor{
					{
						Command: []string{"one", "two"},
					},
				},
			},
			"d41d8cd98f00b204e9800998ecf8427e",
		},
	}

	for _, test := range tests {
		// try the hash a couple times
		for i := 0; i < 3; i++ {
			hash, err := Hash(test.task)
			if err != nil {
				t.Error(err)
			}
			if hash != test.hash {
				t.Errorf("hash mismatch: expected %s got %s", test.hash, hash)
			}
		}
	}
}
