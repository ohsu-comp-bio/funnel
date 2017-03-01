package condor

import (
	"context"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"tes/config"
	pbe "tes/ga4gh"
	"tes/logger"
	sched "tes/scheduler"
	pbr "tes/server/proto"
	"text/template"
	"time"
)

var log = logger.New("condor")

// prefix is a string prefixed to condor worker IDs, so that condor
// workers can be identified by track() below.
const prefix = "condor-"

// NewScheduler returns a new HTCondor Scheduler instance.
func NewScheduler(conf config.Config) sched.Scheduler {
	s := &scheduler{conf}
	go s.track()
	return s
}

type scheduler struct {
	conf config.Config
}

// Schedule schedules a job on the HTCondor queue and returns a corresponding Offer.
func (s *scheduler) Schedule(j *pbe.Job) *sched.Offer {
	log.Debug("Running condor scheduler")

	// TODO could we call condor_submit --dry-run to test if a job would succeed?
	w := &pbr.Worker{
		Id: prefix + sched.GenWorkerID(),
	}
	return sched.NewOffer(w, j, sched.Scores{})
}

// track helps the scheduler know when a job has been assigned to a condor worker,
// so that the worker can be submitted/started via condor. This polls the server
// looking for jobs which are assigned to a worker with a "condor-" ID prefix.
// When such a worker is found, if it has an assigned (inactive) job, a worker is
// started via condor_submit.
func (s *scheduler) track() {
	client, _ := sched.NewClient(s.conf.Worker)
	defer client.Close()

	ticker := time.NewTicker(s.conf.Worker.TrackerRate)

	for {
		<-ticker.C

		resp, err := client.GetWorkers(context.Background(), &pbr.GetWorkersRequest{})
		if err != nil {
			log.Error("Failed GetWorkers request. Recovering.", err)
			continue
		}

		for _, w := range resp.Workers {

			if strings.HasPrefix(w.Id, prefix) &&
				w.State == pbr.WorkerState_Uninitialized {

				w.State = pbr.WorkerState_Initializing
				_, err := client.UpdateWorker(context.TODO(), w)

				if err != nil {
					log.Error("Can't set worker state to initialzing. Skipping start.", err)
					continue
				}

				// TODO there could be an error starting the worker.
				s.startWorker(w.Id)
			}
		}
	}
}

// startWorker submits a job via "condor_submit" to start a new worker.
func (s *scheduler) startWorker(workerID string) {
	log.Debug("Starting condor worker")

	// TODO document that these working dirs need manual cleanup
	workdir := path.Join(s.conf.WorkDir, "condor-scheduler", workerID)
	workdir, _ = filepath.Abs(workdir)
	os.MkdirAll(workdir, 0755)

	w := s.conf.Worker
	w.ID = workerID
	w.Timeout = 0
	w.Storage = s.conf.Storage

	confPath := path.Join(workdir, "worker.conf.yml")
	w.ToYamlFile(confPath)

	workerPath := sched.DetectWorkerPath()

	submitPath := path.Join(workdir, "condor.submit")
	f, _ := os.Create(submitPath)

	submitTpl, _ := template.New("condor.submit").Parse(`
		universe    = vanilla
		executable  = {{.Executable}}
		arguments   = -config worker.conf.yml
		environment = "PATH=/usr/bin"
		log         = {{.WorkDir}}/condor-event-log
		error       = {{.WorkDir}}/tes-worker-stderr
		output      = {{.WorkDir}}/tes-worker-stdout
    input       = {{.Config}}
    should_transfer_files   = YES
    when_to_transfer_output = ON_EXIT
		queue
	`)
	submitTpl.Execute(f, map[string]string{
		"Executable": workerPath,
		"WorkDir":    workdir,
		"Config":     confPath,
	})
	f.Close()

	cmd := exec.Command("condor_submit")
	stdin, _ := os.Open(submitPath)
	cmd.Stdin = stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
