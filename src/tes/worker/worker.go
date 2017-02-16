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

	sched, err := scheduler.NewClient(conf)
	if err != nil {
		return err
	}
	defer sched.Close()

	runners := map[string]*jobRunner{}
	updates := make(updateChan)

	ticker := time.NewTicker(conf.UpdateRate)
	defer ticker.Stop()

	for {
		select {
		// "updates" contains job log, e.g. stdout/err updates.
		case up := <-updates:
			log.Debug("Sending update", "data", up)
			ctx, cleanup := context.WithTimeout(context.Background(), time.Second)
			_, err := sched.UpdateJobLogs(ctx, up)
			if err != nil {
				log.Error("Job log update failed", err)
			}
			cleanup()
			// TODO what if request times out? How to avoid losing update?

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

			// TODO is it possible to get a stale response from the network in gRPC?
			// TODO configurable timeout?
			ctx, cleanup := context.WithTimeout(context.Background(), time.Second)
			resp, err := sched.UpdateWorker(ctx, req)
			cleanup()

			if err != nil {
				log.Error("Couldn't get worker update. Recovering.", err)
				break
			}

			for _, id := range complete {
				delete(runners, id)
			}

			for _, id := range resp.Canceled {
				// Protect against network communication quirks and failures,
				// ensure the job exists.
				if r := runners[id]; r != nil {
					r.Cancel()
				}
			}

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
