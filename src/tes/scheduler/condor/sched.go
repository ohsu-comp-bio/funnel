package condor

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"sync/atomic"
	pbe "tes/ga4gh"
	sched "tes/scheduler"
)

// TODO
const tesBinPath = "/Users/buchanae/projects/task-execution-server/bin/tes-worker"

func NewScheduler(schedAddr string) sched.Scheduler {
	// TODO this should be discovered by watching condor_status/condor_q
	slotCount := int32(10)
	return &scheduler{schedAddr, tesBinPath, slotCount}
}

type scheduler struct {
	schedAddr string
	binPath   string
	// TODO how does the scheduler get up-to-date condor status?
	//      does it call ask for new status on every call to Schedule?
	//      does get status updates every N seconds?
	//      does it listen to a stream of status changes and build a local state?
	//      what happens when the scheduler goes down? How does it rebuild state?
	// TODO this "available" count is a super dumb hack for demo purposes.
	available int32
}

func (s *scheduler) Schedule(t *pbe.Task) sched.Offer {
	log.Println("Running condor scheduler")

	avail := atomic.LoadInt32(&s.available)
	log.Printf("Available: %d", avail)
	if avail == int32(0) {
		return sched.RejectedOffer("Pool is full")
	} else {
		w := sched.Worker{
			ID: sched.GenWorkerID(),
			Resources: sched.Resources{
				// TODO
				CPU:  1,
				RAM:  1.0,
				Disk: 10.0,
			},
		}
		o := sched.NewOffer(t, w)
		go s.observe(o)
		return o
	}
}

func (s *scheduler) observe(o sched.Offer) {
	<-o.Wait()
	if o.Accepted() {
		atomic.AddInt32(&s.available, -1)
		s.startWorker(o.Worker().ID)
		atomic.AddInt32(&s.available, 1)
	} else if o.Rejected() {
		log.Println("Condor offer was rejected")
	}
}

func (s *scheduler) startWorker(workerID string) {
	log.Println("Start condor worker")

	conf := fmt.Sprintf(`
		universe = vanilla
		executable = %s
		arguments = -nworkers 1 -master %s -id %s
		log = log
		error = err
		output = out
		queue
	`, s.binPath, s.schedAddr, workerID)

	log.Printf("Condor submit config: \n%s", conf)

	cmd := exec.Command("condor_submit")
	stdin, _ := cmd.StdinPipe()
	io.WriteString(stdin, conf)
	stdin.Close()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
}
