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

type Config struct {
	MasterAddr string
	Slots      int
	BinPath    string
}

func NewScheduler(c Config) sched.Scheduler {
	return &scheduler{c, int32(c.Slots)}
}

type scheduler struct {
	conf Config
	// TODO this "available" count is a super dumb hack for demo purposes.
	// TODO this should be discovered by watching condor_status/condor_q
	// TODO how does the scheduler get up-to-date condor status?
	//      does it call ask for new status on every call to Schedule?
	//      does get status updates every N seconds?
	//      does it listen to a stream of status changes and build a local state?
	//      what happens when the scheduler goes down? How does it rebuild state?
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
		// TODO there is nothing to actually check/know when a job is finished,
		//      so "available" never gets decremented
	} else if o.Rejected() {
		log.Println("Condor offer was rejected")
	}
}

func (s *scheduler) startWorker(workerID string) {
	log.Println("Start condor worker")

	conf := fmt.Sprintf(`
		universe = vanilla
		executable = %s
		arguments = -numworkers 1 -masteraddr %s -id %s -timeout 0
		environment = "PATH=/usr/bin"
		log = log
		error = err
		output = out
		queue
	`, s.conf.BinPath, s.conf.MasterAddr, workerID)

	log.Printf("Condor submit config: \n%s", conf)

	cmd := exec.Command("condor_submit")
	stdin, _ := cmd.StdinPipe()
	io.WriteString(stdin, conf)
	stdin.Close()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
}
