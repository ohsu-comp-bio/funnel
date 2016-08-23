package ga4gh_taskengine

import (
	"ga4gh-engine/scaling"
	"ga4gh-server/proto"
	"ga4gh-tasks"
	"golang.org/x/net/context"
	//"log"
)

type TaskDB interface {
	ga4gh_task_exec.TaskServiceServer
	ga4gh_task_ref.SchedulerServer
}

func Scheduler(task_server TaskDB, scaler ga4gh_engine_scaling.Scaler) *TaskScheduler {
	return &TaskScheduler{task_server: task_server, scaler: scaler}
}

type TaskScheduler struct {
	task_server TaskDB
	scaler      ga4gh_engine_scaling.Scaler
}

func (self *TaskScheduler) WorkerPing(ctx context.Context, info *ga4gh_task_ref.WorkerInfo) (*ga4gh_task_ref.WorkerInfo, error) {
	self.scaler.PingReceived(info)
	return info, nil
}

func (self *TaskScheduler) GetJobToRun(ctx context.Context, request *ga4gh_task_ref.JobRequest) (*ga4gh_task_ref.JobResponse, error) {
	return self.task_server.GetJobToRun(ctx, request)
}

func (self *TaskScheduler) UpdateJobStatus(ctx context.Context, stat *ga4gh_task_ref.UpdateStatusRequest) (*ga4gh_task_exec.JobId, error) {
	return self.task_server.UpdateJobStatus(ctx, stat)
}

func (self *TaskScheduler) GetQueueInfo(request *ga4gh_task_ref.QueuedTaskInfoRequest, server ga4gh_task_ref.Scheduler_GetQueueInfoServer) error {
	return self.task_server.GetQueueInfo(request, server)
}
