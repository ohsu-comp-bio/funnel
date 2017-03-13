package worker

import (
	"context"
	"tes/config"
	"tes/scheduler"
	pbr "tes/server/proto"
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

func (c *schedClient) UpdateWorker(req *pbr.Worker) (*pbr.UpdateWorkerResponse, error) {
	ctx, cleanup := context.WithTimeout(context.Background(), c.conf.UpdateTimeout)
	resp, err := c.Client.UpdateWorker(ctx, req)
	cleanup()
	return resp, err
}

func (c *schedClient) UpdateJobLogs(up *pbr.UpdateJobLogsRequest) error {
	ctx, cleanup := context.WithTimeout(context.Background(), c.conf.UpdateTimeout)
	_, err := c.Client.UpdateJobLogs(ctx, up)
	cleanup()
	return err
}

func (c *schedClient) Close() {
	c.Client.Close()
}
