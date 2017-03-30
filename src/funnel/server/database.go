package server

import (
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
	"golang.org/x/net/context"
)

// Database represents the interface to the database used by the scheduler, scaler, etc.
// Mostly, this exists so it can be mocked during testing.
type Database interface {
	ReadQueue(n int) []*tes.Job
	AssignJob(*tes.Job, *pbf.Worker)
	CheckWorkers() error
	GetWorkers(context.Context, *pbf.GetWorkersRequest) (*pbf.GetWorkersResponse, error)
	UpdateWorker(context.Context, *pbf.Worker) (*pbf.UpdateWorkerResponse, error)
}
