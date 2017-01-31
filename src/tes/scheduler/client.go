package scheduler

import (
	"context"
	"google.golang.org/grpc"
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
	conn, err := NewRPCConnection(address)
	if err != nil {
    log.Error("Couldn't connect to schduler", err)
		return nil, err
	}

	s := pbr.NewSchedulerClient(conn)
	return &Client{s, conn}, nil
}

// Close closes the client connection.
func (client *Client) Close() {
	client.conn.Close()
}

// SetInitializing sends an UpdateJobStatus request to the scheduler,
// setting the job state to Initializing.
func (client *Client) SetInitializing(ctx context.Context, job *pbe.Job) {
	// Notify the scheduler that the job is running
	client.UpdateJobStatus(ctx,
		&pbr.UpdateStatusRequest{
			Id: job.JobID, State: pbe.State_Initializing})
}

// SetRunning sends an UpdateJobStatus request to the scheduler,
// setting the job state to Running.
func (client *Client) SetRunning(ctx context.Context, job *pbe.Job) {
	// Notify the scheduler that the job is running
	client.UpdateJobStatus(ctx,
		&pbr.UpdateStatusRequest{
			Id: job.JobID, State: pbe.State_Running})
}

// SetFailed sends an UpdateJobStatus request to the scheduler,
// setting the job state to Failed.
func (client *Client) SetFailed(ctx context.Context, job *pbe.Job) {
	// Notify the scheduler that the job failed
	client.UpdateJobStatus(ctx,
		&pbr.UpdateStatusRequest{
			Id: job.JobID, State: pbe.State_Error})
}

// SetComplete sends an UpdateJobStatus request to the scheduler,
// setting the job state to Complete.
func (client *Client) SetComplete(ctx context.Context, job *pbe.Job) {
	// Notify the scheduler that the job is complete
	client.UpdateJobStatus(ctx,
		&pbr.UpdateStatusRequest{
			Id: job.JobID, State: pbe.State_Complete})
}

// PollForJob polls the scheduler for a job assigned to the given worker ID.
func (client *Client) PollForJob(ctx context.Context, workerID string) *pbr.JobResponse {
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
