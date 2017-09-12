package gridengine

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

// name of the scheduler backend
const name = "gridengine"

var log = logger.Sub(name)

// prefix is a string prefixed to gridengine node IDs, so that gridengine
// nodes can be identified by ShouldStartNode() below.
const prefix = name + "-node-"

// NewBackend returns a new grid engine Backend instance.
func NewBackend(conf config.Config) (*Backend, error) {
	return &Backend{
		name:     name,
		conf:     conf,
		template: conf.Backends.GridEngine.Template,
	}, nil
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
	log.Debug("Starting gridengine node")

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
