package scheduler

import (
	"context"
	"google.golang.org/grpc"
	"os"
	"tes/config"
	pbr "tes/server/proto"
	"time"
)

// Client is a client for the scheduler gRPC service.
type Client struct {
	pbr.SchedulerClient
	conn           *grpc.ClientConn
	NewJobPollRate time.Duration
}

// NewClient returns a new Client instance connected to the
// scheduler at a given address (e.g. "localhost:9090")
func NewClient(conf config.Worker) (*Client, error) {
	conn, err := NewRPCConnection(conf.ServerAddress)
	if err != nil {
		log.Error("Couldn't connect to schduler", err)
		return nil, err
	}

	s := pbr.NewSchedulerClient(conn)
	return &Client{s, conn, conf.NewJobPollRate}, nil
}

// Close closes the client connection.
func (client *Client) Close() {
	client.conn.Close()
}

// PollForJob polls the scheduler for a job assigned to the given worker ID.
func (client *Client) PollForJobs(ctx context.Context, workerID string, ch chan<- *pbr.JobResponse) {

  log.Debug("Job poll rate", "rate", client.NewJobPollRate)
	tickChan := time.NewTicker(client.NewJobPollRate).C

	// TODO want ticker that fires immediately
	job := client.RequestJob(ctx, workerID)
	if job != nil {
		ch <- job
	}

	for {
		select {
		case <-ctx.Done():
			return

		case <-tickChan:
			job := client.RequestJob(ctx, workerID)
			if job != nil {
				ch <- job
			}
		}
	}
}

// RequestJob asks the scheduler service for a job. If no job is available, return nil.
func (client *Client) RequestJob(ctx context.Context, workerID string) *pbr.JobResponse {
	hostname, _ := os.Hostname()
	// Ask the scheduler for a task.
	resp, err := client.GetJobToRun(ctx,
		&pbr.JobRequest{
			Worker: &pbr.WorkerInfo{
				Id:       workerID,
				Hostname: hostname,
				// TODO what is last ping for? Why is it the current time?
				LastPing: time.Now().Unix(),
			},
		})

	if err != nil {
		// An error occurred while asking the scheduler for a job.
		// TODO should return error?
		log.Error("Couldn't get job from scheduler", err)

	} else if resp != nil && resp.Job != nil {
		// A job was found
		return resp
	}
	return nil
}
