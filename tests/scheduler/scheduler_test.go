package scheduler

import (
	"github.com/ohsu-comp-bio/funnel/tests"
	"testing"
	"time"
)

func TestReadQueue(t *testing.T) {
	c := tests.DefaultConfig()
	c.Compute = "manual"
	f := tests.NewFunnel(c)
	f.StartServer()

	for i := 0; i < 10; i++ {
		f.Run(`--sh 'echo 1'`)
	}
	time.Sleep(time.Second * 5)

	tasks := f.Scheduler.Queue.ReadQueue(10)

	if len(tasks) != 10 {
		t.Error("unexpected task count", len(tasks))
	}

	min := func(a, b int) int {
		if a < b {
			return a
		}
		return b
	}

	// test that read queue returns tasks in first in first out order
	for i := range tasks {
		j := min(i+1, len(tasks)-1)
		if tasks[i].CreationTime > tasks[j].CreationTime {
			t.Error("unexpected task sort order")
		}
	}
}
