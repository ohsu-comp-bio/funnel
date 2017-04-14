package tests

// This file probably fits better in funnel/server, but due to circular
// imports, it's easier to have it here.

import (
	"funnel/logger"
	pbf "funnel/proto/funnel"
	server_mocks "funnel/server/mocks"
	"testing"
	"time"
)

func init() {
	logger.ForceColors()
}

// Test the simple case of a worker that is alive,
// then doesn't ping in time, and it marked dead
func TestWorkerDead(t *testing.T) {
	conf := server_mocks.NewConfig()
	conf.WorkerPingTimeout = time.Millisecond
	srv := server_mocks.NewServer(conf)
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
	conf := server_mocks.NewConfig()
	conf.WorkerInitTimeout = time.Millisecond
	srv := server_mocks.NewServer(conf)
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

// Test what happens when a worker is marked as dead,
// and then pings in later.
func TestWorkerDeadThenPing(t *testing.T) {
}
