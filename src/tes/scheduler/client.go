package scheduler

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"os"
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
	"time"
)

// Client is a client for the scheduler gRPC service.
type Client struct {
	pbr.SchedulerClient
	conn *grpc.ClientConn
}

// NewClient returns a new Client instance connected to the
// scheduler at a given address (e.g. "localhost:9090")
func NewClient(address string) (*Client, error) {
	conn, err := NewRpcConnection(address)
	if err != nil {
		log.Printf("Error connecting to scheduler: %s", err)
		return nil, err
	}

	s := pbr.NewSchedulerClient(conn)
	return &Client{s, conn}, nil
}

// Close the client connection.
func (client *Client) Close() {
	client.conn.Close()
}

func (client *Client) SetRunning(ctx context.Context, job *pbe.Job) {
	// Notify the scheduler that the job is running
	client.UpdateJobStatus(ctx,
		&pbr.UpdateStatusRequest{
			Id: job.JobID, State: pbe.State_Running})
}

func (client *Client) SetFailed(ctx context.Context, job *pbe.Job) {
	// Notify the scheduler that the job failed
	client.UpdateJobStatus(ctx,
		&pbr.UpdateStatusRequest{
			Id: job.JobID, State: pbe.State_Error})
}

func (client *Client) SetComplete(ctx context.Context, job *pbe.Job) {
	// Notify the scheduler that the job is complete
	client.UpdateJobStatus(ctx,
		&pbr.UpdateStatusRequest{
			Id: job.JobID, State: pbe.State_Complete})
}

func (client *Client) PollForJob(ctx context.Context, workerID string) *pbe.Job {
	// Hard-coding this sleep time because I don't see a need for configuration,
	// but it's easy enough to change that.
	sleep := time.Second * 2
	// "ticker" helps us check for jobs every "sleep" (e.g. 2 seconds).
	ticker := time.NewTicker(sleep)
	defer ticker.Stop()

	job := client.RequestJob(ctx, workerID)
	if job != nil {
		return job
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case <-ticker.C:
			job := client.RequestJob(ctx, workerID)
			if job != nil {
				return job
			}
		}
	}
}

// requestJob asks the scheduler service for a job. If no job is available, return nil.
func (client *Client) RequestJob(ctx context.Context, workerID string) *pbe.Job {
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
		log.Printf("Error getting job from scheduler: %s", err)

	} else if resp != nil && resp.Job != nil {
		// A job was found
		return resp.Job
	}
	return nil
}
