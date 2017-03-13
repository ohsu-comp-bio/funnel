package condor

import (
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
)

var log = logger.New("condor")

// prefix is a string prefixed to condor worker IDs, so that condor
// workers can be identified by ShouldStartWorker() below.
// TODO move to worker metadata to be consistent with GCE
const prefix = "condor-"

// NewScheduler returns a new HTCondor Scheduler instance.
func NewScheduler(conf config.Config) sched.Scheduler {
	return &scheduler{conf}
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

func (s *scheduler) ShouldStartWorker(w *pbr.Worker) bool {
	return strings.HasPrefix(w.Id, prefix) &&
		w.State == pbr.WorkerState_Uninitialized
}

// StartWorker submits a job via "condor_submit" to start a new worker.
func (s *scheduler) StartWorker(w *pbr.Worker) error {
	log.Debug("Starting condor worker")

	// TODO document that these working dirs need manual cleanup
	workdir := path.Join(s.conf.WorkDir, w.Id)
	workdir, _ = filepath.Abs(workdir)
	os.MkdirAll(workdir, 0755)

	c := s.conf.Worker
	c.ID = w.Id
	c.Timeout = 0

	confPath := path.Join(workdir, "worker.conf.yml")
	c.ToYamlFile(confPath)

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

	// TODO better error checking
	return nil
}
