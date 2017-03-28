package scheduler

import (
	pbe "funnel/ga4gh"
	pbr "funnel/server/proto"
	"runtime/debug"
	"testing"
)

func TestDefaultScoresEmptyJob(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DefaultScores panic on empty job/worker\n%s", debug.Stack())
		}
	}()

	j := &pbe.Job{}
	w := &pbr.Worker{}
	DefaultScores(w, j)
}
