package server

import (
	"github.com/ohsu-comp-bio/funnel/config"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests/testutils"
	"testing"
)

// Test a scheduled task is removed from the task queue.
func TestScheduledTaskRemovedFromQueue(t *testing.T) {
	conf := config.DefaultConfig()
	conf = testutils.TempDirConfig(conf)

	// Create database
	db, dberr := NewTaskBolt(conf)
	if dberr != nil {
		t.Fatal("Couldn't open database")
	}

	task := &tes.Task{
		Id: "task-1",
		Executors: []*tes.Executor{
			{
				ImageName: "ubuntu",
				Cmd:       []string{"echo"},
			},
		},
	}

	err := db.QueueTask(task)
	if err != nil {
		t.Fatal("QueueTask failed", err)
	}

	res := db.ReadQueue(10)
	if len(res) != 1 {
		t.Fatal("Expected task in queue")
	}

	err = db.AssignTask(task, &pbs.Node{
		Id: "node-1",
	})
	if err != nil {
		t.Fatal(err)
	}

	res2 := db.ReadQueue(10)
	if len(res2) != 0 {
		t.Fatal("Expected task queue to be empty")
	}
}
