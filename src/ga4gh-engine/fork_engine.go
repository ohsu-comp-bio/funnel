
package ga4gh_taskengine

import (
  	"time"
  	"log"
  	"sync/atomic"
	context "golang.org/x/net/context"
	"ga4gh-server/proto"
	"os"
	"ga4gh-tasks"
	//proto "github.com/golang/protobuf/proto"
)


type ForkManager struct {
	procCount int
	running bool
	files FileMapper
	sched ga4gh_task_ref.SchedulerClient
	workdir string
	workerId string
	ctx context.Context
	check_func func(status EngineStatus)
	status EngineStatus
}



func (self *ForkManager) worker(inchan chan ga4gh_task_exec.TaskOp) {
  for job := range inchan {
    atomic.AddInt32(&self.status.ActiveJobs, 1)
    atomic.AddInt32(&self.status.JobCount, 1)
    log.Printf("Launch job: %s", job)
    s := ga4gh_task_exec.State_Running
    self.sched.UpdateTaskOpStatus(self.ctx, &ga4gh_task_ref.UpdateStatusRequest{Id:job.TaskOpId, State:s})
    err := RunJob(&job, self.files)
  	if err != nil {
		log.Printf("Job error: %s", err)
		self.sched.UpdateTaskOpStatus(self.ctx, &ga4gh_task_ref.UpdateStatusRequest{Id:job.TaskOpId, State:ga4gh_task_exec.State_Error})
	} else {
		self.sched.UpdateTaskOpStatus(self.ctx, &ga4gh_task_ref.UpdateStatusRequest{Id:job.TaskOpId, State:ga4gh_task_exec.State_Complete})
	}
    atomic.AddInt32(&self.status.ActiveJobs, -1)
  }
}

func (self *ForkManager) watcher(sched ga4gh_task_ref.SchedulerClient, filestore FileMapper) {
  self.sched = sched
  self.files = filestore
  hostname, _ := os.Hostname()
  jobchan := make(chan ga4gh_task_exec.TaskOp, 10)
  for i := 0; i < self.procCount; i++ {
    go self.worker(jobchan)
  }
  var sleep_size int64 = 1
  for self.running {
    if self.check_func != nil {
      self.check_func(self.status)
    }
    job, err := self.sched.GetJobToRun(self.ctx,
      &ga4gh_task_ref.JobRequest{
        Worker: &ga4gh_task_ref.WorkerInfo{
          Id:self.workerId,
          Hostname:hostname,
          LastPing:time.Now().Unix(),
        },
      })
	if err != nil {
		log.Print(err)
	}
    if job != nil && job.Task != nil {
      sleep_size = 1
      log.Printf("Found job: %s", job)
      jobchan <- *job.Task
    } else {
      log.Printf("No jobs found")
      if (sleep_size < 20) {
        //  sleep_size += 1
      }
      time.Sleep(time.Second * time.Duration(sleep_size))
    }
  }
  close(jobchan)
}

func (self *ForkManager) Start(engine ga4gh_task_ref.SchedulerClient, files FileMapper) {
  go self.watcher(engine, files)
}

func (self *ForkManager) Run(engine ga4gh_task_ref.SchedulerClient, files FileMapper) {
  self.watcher(engine, files)
}

func (self *ForkManager) SetStatusCheck( check_func func(status EngineStatus)) {
  self.check_func = check_func
}

func NewLocalManager(procCount int, workdir string, workerId string) (*ForkManager, error) {
  return &ForkManager{
    procCount:procCount,
    running:true,
    workdir:workdir,
    workerId:workerId,
    ctx:context.Background(),
  }, nil
}