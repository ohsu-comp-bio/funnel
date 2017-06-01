package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/scheduler"
)

// Defines some helpers for RPC calls in the code above
type schedClient struct {
	scheduler.Client
	conf config.Worker
}

func newSchedClient(conf config.Worker) (*schedClient, error) {
	sched, err := scheduler.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return &schedClient{sched, conf}, nil
}

func (c *schedClient) UpdateWorker(req *pbf.Worker) (*pbf.UpdateWorkerResponse, error) {
	ctx, cleanup := context.WithTimeout(context.Background(), c.conf.UpdateTimeout)
	resp, err := c.Client.UpdateWorker(ctx, req)
	cleanup()
	return resp, err
}

func (c *schedClient) UpdateExecutorLogs(up *pbf.UpdateExecutorLogsRequest) error {
	ctx, cleanup := context.WithTimeout(context.Background(), c.conf.UpdateTimeout)
	_, err := c.Client.UpdateExecutorLogs(ctx, up)
	cleanup()
	return err
}

func (c *schedClient) UpdateTaskLogs(up *pbf.UpdateTaskLogsRequest) error {
	ctx, cleanup := context.WithTimeout(context.Background(), c.conf.UpdateTimeout)
	_, err := c.Client.UpdateTaskLogs(ctx, up)
	cleanup()
	return err
}

func (c *schedClient) Close() {
	c.Client.Close()
}
