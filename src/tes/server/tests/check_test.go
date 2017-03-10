package tests

// This file probably fits better in tes/server, but due to circular
// imports, it's easier to have it here.

import (
	"tes/config"
	"tes/logger"
	server_mocks "tes/server/mocks"
	pbr "tes/server/proto"
	"testing"
	"time"
)

func init() {
	logger.ForceColors()
}

// Test the simple case of a worker that is alive,
// then doesn't ping in time, and it marked dead
func TestWorkerDead(t *testing.T) {
	conf := config.DefaultConfig()
	conf.WorkerPingTimeout = time.Millisecond
	srv := server_mocks.MockServerFromConfig(conf)
	defer srv.Close()

	srv.AddWorker(&pbr.Worker{
		Id:    "test-worker",
		State: pbr.WorkerState_Alive,
	})

	time.Sleep(conf.WorkerPingTimeout * 2)
	srv.DB.CheckWorkers()

	workers := srv.GetWorkers()
	if workers[0].State != pbr.WorkerState_Dead {
		t.Error("Expected worker to be dead")
	}
}

// Test what happens when a worker never starts.
// It should be marked as dead.
func TestWorkerInitFail(t *testing.T) {
	conf := config.DefaultConfig()
	conf.WorkerInitTimeout = time.Millisecond
	srv := server_mocks.MockServerFromConfig(conf)
	defer srv.Close()

	srv.AddWorker(&pbr.Worker{
		Id:    "test-worker",
		State: pbr.WorkerState_Initializing,
	})

	time.Sleep(conf.WorkerInitTimeout * 2)
	srv.DB.CheckWorkers()
	workers := srv.GetWorkers()

	if workers[0].State != pbr.WorkerState_Dead {
		t.Error("Expected worker to be dead")
	}
}

// Test what happens when a worker is marked as dead,
// and then pings in later.
func TestWorkerDeadThenPing(t *testing.T) {
}
