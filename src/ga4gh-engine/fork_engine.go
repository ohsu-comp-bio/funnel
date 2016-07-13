
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
	workerId string
	ctx context.Context
	check_func func(status EngineStatus)
	status EngineStatus
}



func (self *ForkManager) worker(inchan chan ga4gh_task_exec.Job) {
  for job := range inchan {
    atomic.AddInt32(&self.status.ActiveJobs, 1)
    atomic.AddInt32(&self.status.JobCount, 1)
    log.Printf("Launch job: %s", job)
    s := ga4gh_task_exec.State_Running
    self.sched.UpdateJobStatus(self.ctx, &ga4gh_task_ref.UpdateStatusRequest{Id:job.JobId, State:s})
    err := RunJob(&job, self.files)
  	if err != nil {
		log.Printf("Job error: %s", err)
		self.sched.UpdateJobStatus(self.ctx, &ga4gh_task_ref.UpdateStatusRequest{Id:job.JobId, State:ga4gh_task_exec.State_Error})
	} else {
		self.sched.UpdateJobStatus(self.ctx, &ga4gh_task_ref.UpdateStatusRequest{Id:job.JobId, State:ga4gh_task_exec.State_Complete})
	}
    atomic.AddInt32(&self.status.ActiveJobs, -1)
  }
}

func (self *ForkManager) watcher(sched ga4gh_task_ref.SchedulerClient, filestore FileMapper) {
  self.sched = sched
  self.files = filestore
  hostname, _ := os.Hostname()
  jobchan := make(chan ga4gh_task_exec.Job, 10)
  for i := 0; i < self.procCount; i++ {
    go self.worker(jobchan)
  }
  var sleep_size int64 = 1
  for self.running {
    if self.check_func != nil {
      self.check_func(self.status)
    }
    task, err := self.sched.GetJobToRun(self.ctx,
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
    if task != nil && task.Job != nil {
      sleep_size = 1
      log.Printf("Found job: %s", task)
      jobchan <- *task.Job
    } else {
      //log.Printf("No jobs found")
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

func NewLocalManager(procCount int, workerId string) (*ForkManager, error) {
  return &ForkManager{
    procCount:procCount,
    running:true,
    workerId:workerId,
    ctx:context.Background(),
  }, nil
}