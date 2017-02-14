package scheduler

import (
	uuid "github.com/nu7hatch/gouuid"
	pbe "tes/ga4gh"
	server "tes/server"
	"time"
)

// Scheduler is responsible for scheduling a job. It has a single method which
// is responsible for taking a job and returning an Offer which describes whether
// a scheduler can run the job, how many resources it can offer, and anything that
// might allow a central scheduler to decide where best to run the job.
//
// For example, a system might have a separate scheduler for each of
// Google Cloud, AWS, and on-premise HTCondor clusters. For a given job,
// the Google Cloud and AWS schedulers might determine that the job cannot be run
// (maybe due to data locality restrictions) and they return rejected Offers, while
// the HTCondor returns an accepted Offer. A central scheduler can then determine
// that the job should be assigned to the HTCondor cluster.
type Scheduler interface {
	Schedule(*pbe.Job) *Worker
}

type fitPred func(*pbe.Job, *Worker) bool

// StartScheduling starts a scheduling loop, pulling 10 jobs from the database,
// and sending those to the given scheduler. This happens every 5 seconds.
func StartScheduling(db *server.TaskBolt, sched Scheduler, pollRate time.Duration) {
	tickChan := time.NewTicker(pollRate).C

	for {
		<-tickChan
		for _, job := range db.ReadQueue(10) {
			worker := sched.Schedule(job)
			if worker != nil {
				log.Debug("Assigning job to worker",
					"jobID", job.JobID,
					"workerID", worker.ID,
				)
				db.AssignJob(job.JobID, worker.ID)
			} else {
				log.Info("No worker could be scheduled for job", "jobID", job.JobID)
			}
		}
	}
}

// GenWorkerID returns a UUID string.
func GenWorkerID() string {
	u, _ := uuid.NewV4()
	return "worker-" + u.String()
}

func ResourcesFit(j *pbe.Job, w *Worker) bool {
	req := j.Task.Resources

	// If the task didn't include resource requirements,
	// assume it fits.
	//
	// TODO think about whether this is the desired behavior
	if req == nil {
		return true
	}
	switch {
	case w.Available.CPUs < req.MinimumCpuCores:
		return false
	case w.Available.RAM < req.MinimumRamGb:
		return false
		// TODO check volumes
	}
	return true
}

// TODO should have a predicate which understands authorization
//      - storage
//      - other auth resources?
//      - does storage need to be scheduler specific?

// TODO other predicate ideas
// - preemptible
// - zones
// - port checking
// - disk conflict
// - labels/selectors
// - host name

func Fit(job *pbe.Job, workers []*Worker) []*Worker {
	fit := []*Worker{}
	predicates := []fitPred{
		ResourcesFit,
	}

	for _, w := range workers {
		fits := true
		for _, pred := range predicates {
			if ok := pred(job, w); !ok {
				fits = false
				break
			}
		}
		if fits {
			fit = append(fit, w)
		}
	}
	return fit
}
