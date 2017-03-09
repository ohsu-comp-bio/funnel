package server

import (
	"golang.org/x/net/context"
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
)

type Database interface {
	ReadQueue(n int) []*pbe.Job
	AssignJob(*pbe.Job, *pbr.Worker)
	CheckWorkers() error
	GetWorkers(context.Context, *pbr.GetWorkersRequest) (*pbr.GetWorkersResponse, error)
	UpdateWorker(context.Context, *pbr.Worker) (*pbr.UpdateWorkerResponse, error)
}
