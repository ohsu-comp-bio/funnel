package server

import (
	pbe "funnel/ga4gh"
	pbr "funnel/server/proto"
	"golang.org/x/net/context"
)

// Database represents the interface to the database used by the scheduler, scaler, etc.
// Mostly, this exists so it can be mocked during testing.
type Database interface {
	ReadQueue(n int) []*pbe.Job
	AssignJob(*pbe.Job, *pbr.Worker)
	CheckWorkers() error
	GetWorkers(context.Context, *pbr.GetWorkersRequest) (*pbr.GetWorkersResponse, error)
	UpdateWorker(context.Context, *pbr.Worker) (*pbr.UpdateWorkerResponse, error)
}
