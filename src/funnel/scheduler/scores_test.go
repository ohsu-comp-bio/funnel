package scheduler

import (
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"runtime/debug"
	"testing"
)

func TestDefaultScoresEmptyTask(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DefaultScores panic on empty task/worker\n%s", debug.Stack())
		}
	}()

	j := &tes.Task{}
	w := &pbf.Worker{}
	DefaultScores(w, j)
}
