package pbs

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"os"
	"os/exec"
	"strings"
)

// Name of the scheduler backend
const Name = "pbs"

var log = logger.Sub(Name)

// prefix is a string prefixed to pbs node IDs, so that pbs
// nodes can be identified by ShouldStartNode() below.
const prefix = "pbs-node-"

// NewBackend returns a new PBS Backend instance.
func NewBackend(conf config.Config) (scheduler.Backend, error) {
	return &Backend{
		name:     Name,
		conf:     conf,
		template: conf.Backends.PBS.Template,
	}, nil
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
	return scheduler.SetupSingleTaskNode(prefix, s.conf.Scheduler.Node, t)
}

// ShouldStartNode is part of the Scaler interface and returns true
// when the given node needs to be started by Backend.StartNode
func (s *Backend) ShouldStartNode(w *pbs.Node) bool {
	return strings.HasPrefix(w.Id, prefix) &&
		w.State == pbs.NodeState_UNINITIALIZED
}

// StartNode submits a task via "sbatch" to start a new node.
func (s *Backend) StartNode(w *pbs.Node) error {
	log.Debug("Starting pbs node")

	submitPath, err := scheduler.SetupTemplatedHPCNode(s.name, s.template, s.conf, w)
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
