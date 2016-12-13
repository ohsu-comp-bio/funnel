package slot

import (
	"fmt"
	"log"
	"os"
	"sync"
	pbe "tes/ga4gh"
	"tes/server"
	pbr "tes/server/proto"
	"tes/worker"
	"time"
)

// Used by slots to communicate their status with the pool.
type SlotStatus int32

const (
	Idle SlotStatus = iota
	Running
)

type SlotId string

// A slot requests a job from the scheduler, runs it, and repeats.
// There are many slots in a pool, which provides job concurrency.
type Slot struct {
	Id SlotId
	// Address of the scheduler service, e.g. localhost:9090
	schedulerAddress string
	// sleepDuration controls how long to sleep when no jobs are available
	sleepDuration time.Duration
	// File system configuration, to be passed to the worker engine.
	fileConfig FileConfig
	status     SlotStatus
	statusMtx  sync.Mutex
}

func NewDefaultSlot(id SlotId, schedAddr string, fileConf FileConfig) *Slot {
	return &Slot{
		Id:               id,
		schedulerAddress: schedAddr,
		sleepDuration:    time.Second * 2,
		fileConfig:       fileConf,
	}
}

func (slot *Slot) Status() (status SlotStatus) {
	slot.statusMtx.Lock()
	status = slot.status
	slot.statusMtx.Unlock()
	return
}

func (slot *Slot) setStatus(status SlotStatus) {
	slot.statusMtx.Lock()
	slot.status = status
	slot.statusMtx.Unlock()
}

// Start processing jobs. Ask the scheduler for a job, run it,
// send status updates to the pool/log/scheduler, and repeat.
func (slot *Slot) Start(ctx Context) {
	log.Printf("Starting slot: %s", slot.Id)

	// Get a client for the scheduler service
	sched, err := tes_server.NewSchedulerClient(slot.schedulerAddress)
	defer sched.Close()
	if err != nil {
		log.Printf("Error connecting to scheduler: %s", err)
		log.Printf("Closing slot")
		return
	}

	// ticker helps us check for jobs every slot.sleepDuration
	ticker := time.NewTicker(slot.sleepDuration)
	defer ticker.Stop()

	job := requestJob(ctx, sched, slot.Id)

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			job = requestJob(ctx, sched, slot.Id)

			if job != nil {
				log.Printf("Got job: %s", job.JobID)
				// TODO the engine should probably be responsible for updating job status
				//      Then the slot's only responsibility is to communicate availability
				//      with the scheduler and pool?
				//setRunning(ctx, sched, job)
				slot.setStatus(Running)

				err := slot.runJob(sched, job)
				if err != nil {
					log.Printf("Failed to run job [%s]: %s", job.JobID, err)
				}
			}

			slot.setStatus(Idle)
		}
	}
}

// Run a job.
func (slot *Slot) runJob(sched *tes_server.SchedulerClient, job *pbe.Job) error {
	fileClient := getFileClient(slot.fileConfig)
	fileMapper := tesTaskEngineWorker.NewFileMapper(fileClient, slot.fileConfig.VolumeDir)

	// TODO should RunJob get a context?
	return tesTaskEngineWorker.RunJob(sched, job, *fileMapper)
}

// Generate a slot ID based on a pool ID + slot number.
func GenSlotId(poolId PoolId, i int) SlotId {
	return SlotId(fmt.Sprintf("%s-%d", poolId, i))
}

// requestJob asks the scheduler service for a job. If no job is available, return nil.
func requestJob(ctx Context, sched *tes_server.SchedulerClient, id SlotId) *pbe.Job {
	hostname, _ := os.Hostname()
	// Ask the scheduler for a task.
	resp, err := sched.GetJobToRun(ctx,
		&pbr.JobRequest{
			Worker: &pbr.WorkerInfo{
				Id:       string(id),
				Hostname: hostname,
				// TODO what is last ping for? Why is it the current time?
				LastPing: time.Now().Unix(),
			},
		})

	if err != nil {
		// An error occurred while asking the scheduler for a job.
		log.Printf("Error getting job from scheduler: %s", err)

	} else if resp != nil && resp.Job != nil {
		// A job was found
		return resp.Job
	}
	return nil
}
