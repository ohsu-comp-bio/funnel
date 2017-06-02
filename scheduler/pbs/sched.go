package pbs

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"os"
	"os/exec"
	"strings"
)

var log = logger.New("pbs")

// prefix is a string prefixed to pbs worker IDs, so that pbs
// workers can be identified by ShouldStartWorker() below.
const prefix = "pbs-worker-"

// Plugin provides the PBS scheduler backend plugin.
var Plugin = &scheduler.BackendPlugin{
	Name:   "pbs",
	Create: NewBackend,
}

// NewBackend returns a new PBS Backend instance.
func NewBackend(conf config.Config) (scheduler.Backend, error) {
	b := scheduler.Backend(&Backend{
		name:     "pbs",
		conf:     conf,
		template: conf.Backends.PBS.Template,
	})
	return b, nil
}

// Backend represents the PBS backend.
type Backend struct {
	name     string
	conf     config.Config
	template string
}

// Schedule schedules a task on the PBS queue and returns a corresponding Offer.
func (s *Backend) Schedule(t *tes.Task) *scheduler.Offer {
	log.Debug("Running pbs scheduler")
	return scheduler.ScheduleSingleTaskWorker(prefix, s.conf.Worker, t)
}

// ShouldStartWorker is part of the Scaler interface and returns true
// when the given worker needs to be started by Backend.StartWorker
func (s *Backend) ShouldStartWorker(w *pbf.Worker) bool {
	return strings.HasPrefix(w.Id, prefix) &&
		w.State == pbf.WorkerState_UNINITIALIZED
}

// StartWorker submits a task via "sbatch" to start a new worker.
func (s *Backend) StartWorker(w *pbf.Worker) error {
	log.Debug("Starting pbs worker")

	submitPath, err := scheduler.SetupTemplatedHPCWorker(s.name, s.template, s.conf, w)
	if err != nil {
		return err
	}

	cmd := exec.Command("qsub", submitPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
