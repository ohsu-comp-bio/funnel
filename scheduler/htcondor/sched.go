package htcondor

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

// Name of the scheduler backend
const Name = "htcondor"

var log = logger.Sub(Name)

// prefix is a string prefixed to condor worker IDs, so that condor
// workers can be identified by ShouldStartWorker() below.
const prefix = "htcondor-worker-"

// NewBackend returns a new HTCondor Backend instance.
func NewBackend(conf config.Config) (scheduler.Backend, error) {
	b := scheduler.Backend(&Backend{
		name:     "htcondor",
		conf:     conf,
		template: conf.Backends.HTCondor.Template,
	})
	return b, nil
}

// Backend represents the HTCondor backend.
type Backend struct {
	name     string
	conf     config.Config
	template string
}

// Schedule schedules a task on the HTCondor queue and returns a corresponding Offer.
func (s *Backend) Schedule(t *tes.Task) *scheduler.Offer {
	log.Debug("Running htcondor scheduler")
	return scheduler.ScheduleSingleTaskWorker(prefix, s.conf.Worker, t)
}

// ShouldStartWorker is part of the Scaler interface and returns true
// when the given worker needs to be started by Backend.StartWorker
func (s *Backend) ShouldStartWorker(w *pbf.Worker) bool {
	return strings.HasPrefix(w.Id, prefix) &&
		w.State == pbf.WorkerState_UNINITIALIZED
}

// StartWorker submits a task via "condor_submit" to start a new worker.
func (s *Backend) StartWorker(w *pbf.Worker) error {
	log.Debug("Starting htcondor worker")

	submitPath, err := scheduler.SetupTemplatedHPCWorker(s.name, s.template, s.conf, w)
	if err != nil {
		return err
	}

	cmd := exec.Command("condor_submit", submitPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
