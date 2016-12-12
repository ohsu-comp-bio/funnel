package pool

import (
	pbe "tes/ga4gh"
	"tes/server"
	pbr "tes/server/proto"
)

/****************************************************
 * Helpers for verbose scheduler client code.
 ***************************************************/

func setRunning(ctx Context, sched *tes_server.SchedulerClient, job pbe.Job) {
	// Notify the scheduler that the job is running
	sched.UpdateJobStatus(ctx,
		&pbr.UpdateStatusRequest{
			Id: job.JobID, State: pbe.State_Running})
}

func setFailed(ctx Context, sched *tes_server.SchedulerClient, job pbe.Job) {
	// Notify the scheduler that the job failed
	sched.UpdateJobStatus(ctx,
		&pbr.UpdateStatusRequest{
			Id: job.JobID, State: pbe.State_Error})
}

func setComplete(ctx Context, sched *tes_server.SchedulerClient, job pbe.Job) {
	// Notify the scheduler that the job is complete
	sched.UpdateJobStatus(ctx,
		&pbr.UpdateStatusRequest{
			Id: job.JobID, State: pbe.State_Complete})
}
