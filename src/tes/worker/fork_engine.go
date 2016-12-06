package tesTaskEngineWorker

import (
	context "golang.org/x/net/context"
	"log"
	"os"
	"sync/atomic"
	"tes/ga4gh"
	"tes/server/proto"
	"time"
	//proto "github.com/golang/protobuf/proto"
)

// ForkManager documentation
// TODO: documentation
type ForkManager struct {
	procCount int
	running   bool
	files     FileMapper
	sched     ga4gh_task_ref.SchedulerClient
	workerID  string
	ctx       context.Context
	checkFunc func(status EngineStatus)
	status    EngineStatus
}

func (forkManager *ForkManager) worker(inchan chan ga4gh_task_exec.Job) {
	for job := range inchan {
		atomic.AddInt32(&forkManager.status.ActiveJobs, 1)
		atomic.AddInt32(&forkManager.status.JobCount, 1)
		log.Printf("Launch job: %s", job)
		s := ga4gh_task_exec.State_Running
		forkManager.sched.UpdateJobStatus(forkManager.ctx, &ga4gh_task_ref.UpdateStatusRequest{Id: job.JobID, State: s})
		err := RunJob(&job, forkManager.files)
		if err != nil {
			log.Printf("Job error: %s", err)
			forkManager.sched.UpdateJobStatus(forkManager.ctx, &ga4gh_task_ref.UpdateStatusRequest{Id: job.JobID, State: ga4gh_task_exec.State_Error})
		} else {
			forkManager.sched.UpdateJobStatus(forkManager.ctx, &ga4gh_task_ref.UpdateStatusRequest{Id: job.JobID, State: ga4gh_task_exec.State_Complete})
		}
		atomic.AddInt32(&forkManager.status.ActiveJobs, -1)
	}
}

func (forkManager *ForkManager) watcher(sched ga4gh_task_ref.SchedulerClient, filestore FileMapper) {
	forkManager.sched = sched
	forkManager.files = filestore
	hostname, _ := os.Hostname()
	jobchan := make(chan ga4gh_task_exec.Job, 10)
	for i := 0; i < forkManager.procCount; i++ {
		go forkManager.worker(jobchan)
	}
	var sleepSize int64 = 1
	for forkManager.running {
		if forkManager.checkFunc != nil {
			forkManager.checkFunc(forkManager.status)
		}
		task, err := forkManager.sched.GetJobToRun(forkManager.ctx,
			&ga4gh_task_ref.JobRequest{
				Worker: &ga4gh_task_ref.WorkerInfo{
					Id:       forkManager.workerID,
					Hostname: hostname,
					LastPing: time.Now().Unix(),
				},
			})
		if err != nil {
			log.Print(err)
		}
		if task != nil && task.Job != nil {
			sleepSize = 1
			log.Printf("Found job: %s", task)
			jobchan <- *task.Job
		} else {
			//log.Printf("No jobs found")
			if sleepSize < 20 {
				//  sleepSize += 1
			}
			time.Sleep(time.Second * time.Duration(sleepSize))
		}
	}
	close(jobchan)
}

// Start documentation
// TODO: documentation
func (forkManager *ForkManager) Start(engine ga4gh_task_ref.SchedulerClient, files FileMapper) {
	go forkManager.watcher(engine, files)
}

// Run documentation
// TODO: documentation
func (forkManager *ForkManager) Run(engine ga4gh_task_ref.SchedulerClient, files FileMapper) {
	forkManager.watcher(engine, files)
}

// SetStatusCheck documentation
// TODO: documentation
func (forkManager *ForkManager) SetStatusCheck(checkFunc func(status EngineStatus)) {
	forkManager.checkFunc = checkFunc
}

// NewLocalManager documentation
// TODO: documentation
func NewLocalManager(procCount int, workerID string) (*ForkManager, error) {
	return &ForkManager{
		procCount: procCount,
		running:   true,
		workerID:  workerID,
		ctx:       context.Background(),
	}, nil
}
