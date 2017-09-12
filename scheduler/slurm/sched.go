package slurm

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
const name = "slurm"

var log = logger.Sub(name)

// prefix is a string prefixed to slurm node IDs, so that slurm
// nodes can be identified by ShouldStartNode() below.
const prefix = name + "-node-"

// NewBackend returns a new SLURM Backend instance.
func NewBackend(conf config.Config) (*Backend, error) {
	return &Backend{
		name:     name,
		conf:     conf,
		template: conf.Backends.SLURM.Template,
	}, nil
}

// Backend represents the SLURM backend.
type Backend struct {
	name     string
	conf     config.Config
	template string
}

// Schedule schedules a task on the SLURM queue and returns a corresponding Offer.
func (s *Backend) Schedule(t *tes.Task) *scheduler.Offer {
	log.Debug("Running slurm scheduler")
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
	log.Debug("Starting slurm node")

	submitPath, err := scheduler.SetupTemplatedHPCNode(s.name, s.template, s.conf, w)
	if err != nil {
		return err
	}

	cmd := exec.Command("sbatch", submitPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
