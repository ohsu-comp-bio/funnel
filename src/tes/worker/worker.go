package worker

import (
	"context"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	"tes/scheduler"
	pbr "tes/server/proto"
	"time"
)

type updateChan chan *pbr.UpdateJobLogsRequest

func Run(conf config.Worker) error {
	log := logger.New("worker", "workerID", conf.ID)
	log.Info("Running worker")
	defer log.Info("Shutting down")

	sched, err := newSchedClient(conf)
	if err != nil {
		return err
	}

	defer func() {
		// Tell the scheduler that the worker is gone.
		sched.WorkerGone()
	}()
	defer sched.Close()

	// Tracks active job runners
	runners := map[string]*jobRunner{}
	// Allows job/step runners to send log updates
	updates := make(updateChan)

	// Ticker controls how often the worker make an UpdateWorker() RPC
	ticker := time.NewTicker(conf.UpdateRate)
	defer ticker.Stop()

	for {
		select {
		// "updates" contains job log, e.g. stdout/err updates.
		case up := <-updates:
			// UpdateJobLogs() is more lightweight than UpdateWorker(),
			// which is why it happens separately and at a different rate.
			err := sched.UpdateJobLogs(up)
			if err != nil {
				// TODO if the request failed, the job update is lost and the logs
				//      are corrupted. Cache logs to prevent this?
				log.Error("Job log update failed", err)
			}

		// Ping the server every tick and receive updates,
		// including new jobs and canceled jobs.
		case <-ticker.C:
			req := &pbr.UpdateWorkerRequest{
				Id: conf.ID,
				// TODO how does this even build?
				Resources: conf.Resources,
				// TODO
				Hostname: "unknown",
				States:   map[string]pbe.State{},
			}
			complete := []string{}

			for id, runner := range runners {
				state := runner.State()
				switch state {
				case pbe.State_Canceled, pbe.State_Error, pbe.State_Complete:
					req.States[id] = state
					complete = append(complete, id)

				case pbe.State_Running, pbe.State_Initializing:
					req.States[id] = state
				default:
					log.Error("Unexpected job runner state. Defaulting to Initialzing",
						"state", state)
					req.States[id] = pbe.State_Initializing
				}
			}

			resp, err := sched.UpdateWorker(req)
			if err != nil {
				log.Error("Couldn't get worker update. Recovering.", err)
				break
			}

			// Clean up tracked job runners from jobs that are complete
			for _, id := range complete {
				delete(runners, id)
			}

			// If the server sent "cancel" signals for jobs, call runner.Cancel()
			for _, id := range resp.Canceled {
				// Protect against network communication quirks and failures,
				// ensure the job exists.
				if r := runners[id]; r != nil {
					r.Cancel()
				}
			}

			// Start new job runners for any assigned jobs.
			for _, a := range resp.GetAssigned() {
				log.Debug("Worker received assignment", "assignment", a)
				// Protect against network communication quirks and failures,
				// ensure the job only gets started once.
				id := a.Job.JobID
				if runners[id] == nil {
					// Start the job runner
					r := newJobRunner(conf, a, updates)
					runners[id] = r
					go r.Run()
				}
			}
		}
	}
}

// Defines some helpers for RPC calls in the code above
type schedClient struct {
	*scheduler.Client
	conf config.Worker
}

func newSchedClient(conf config.Worker) (*schedClient, error) {
	sched, err := scheduler.NewClient(conf)
	if err != nil {
		return nil, err
	}
	return &schedClient{sched, conf}, nil
}

func (c *schedClient) UpdateWorker(req *pbr.UpdateWorkerRequest) (*pbr.UpdateWorkerResponse, error) {
	// TODO is it possible to get a stale response from the network in gRPC?
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

func (c *schedClient) WorkerGone() {
	ctx, cleanup := context.WithTimeout(context.Background(), c.conf.UpdateTimeout)
	// Errors are ignored because the worker is shutting down anyway
	c.Client.WorkerGone(ctx, &pbr.WorkerGoneRequest{
		Id: c.conf.ID,
	})
	cleanup()
}
