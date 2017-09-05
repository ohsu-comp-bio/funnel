package scheduler

import (
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"runtime/debug"
	"testing"
)

func TestDefaultScoresEmptyTask(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("DefaultScores panic on empty task/node\n%s", debug.Stack())
		}
	}()

	j := &tes.Task{}
	w := &pbs.Node{}
	DefaultScores(w, j)
}
