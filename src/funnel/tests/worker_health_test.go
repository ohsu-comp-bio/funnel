package tests

import (
	pbf "funnel/proto/funnel"
	"testing"
	"time"
)

// Test the simple case of a worker that is alive,
// then doesn't ping in time, and it marked dead
func TestWorkerDead(t *testing.T) {
	conf := NewConfig()
	conf.WorkerPingTimeout = time.Millisecond
	srv := NewFunnel(conf)
	defer srv.Stop()

	srv.AddWorker(&pbf.Worker{
		Id:    "test-worker",
		State: pbf.WorkerState_Alive,
	})

	time.Sleep(conf.WorkerPingTimeout * 2)
	srv.DB.CheckWorkers()

	workers := srv.GetWorkers()
	if workers[0].State != pbf.WorkerState_Dead {
		t.Error("Expected worker to be dead")
	}
}

// Test what happens when a worker never starts.
// It should be marked as dead.
func TestWorkerInitFail(t *testing.T) {
	conf := NewConfig()
	conf.WorkerInitTimeout = time.Millisecond
	srv := NewFunnel(conf)
	defer srv.Stop()

	srv.AddWorker(&pbf.Worker{
		Id:    "test-worker",
		State: pbf.WorkerState_Initializing,
	})

	time.Sleep(conf.WorkerInitTimeout * 2)
	srv.DB.CheckWorkers()
	workers := srv.GetWorkers()

	if workers[0].State != pbf.WorkerState_Dead {
		t.Error("Expected worker to be dead")
	}
}
