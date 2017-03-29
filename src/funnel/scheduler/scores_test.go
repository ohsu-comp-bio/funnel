package scheduler

import (
	tes "funnel/proto/tes"
	pbf "funnel/proto/funnel"
	"runtime/debug"
	"testing"
)

func TestDefaultScoresEmptyJob(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DefaultScores panic on empty job/worker\n%s", debug.Stack())
		}
	}()

	j := &tes.Job{}
	w := &pbf.Worker{}
	DefaultScores(w, j)
}
