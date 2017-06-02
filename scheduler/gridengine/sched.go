package gridengine

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

var log = logger.New("gridengine")

// prefix is a string prefixed to gridengine worker IDs, so that gridengine
// workers can be identified by ShouldStartWorker() below.
const prefix = "gridengine-worker-"

// Plugin provides the grid engine scheduler backend plugin.
var Plugin = &scheduler.BackendPlugin{
	Name:   "gridengine",
	Create: NewBackend,
}

// NewBackend returns a new grid engine Backend instance.
func NewBackend(conf config.Config) (scheduler.Backend, error) {
	b := scheduler.Backend(&Backend{
		name:     "gridengine",
		conf:     conf,
		template: conf.Backends.GridEngine.Template,
	})
	return b, nil
}

// Backend represents the grid engine backend.
type Backend struct {
	name     string
	conf     config.Config
	template string
}

// Schedule schedules a task on the grid engine queue and returns a corresponding Offer.
func (s *Backend) Schedule(t *tes.Task) *scheduler.Offer {
	log.Debug("Running gridengine scheduler")
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
	log.Debug("Starting gridengine worker")

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
