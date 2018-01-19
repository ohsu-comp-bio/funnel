package slurm

import (
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
)

// NewBackend returns a new Slurm HPCBackend instance.
func NewBackend(conf config.Config) *compute.HPCBackend {
	return compute.NewHPCBackend("slurm", "sbatch", conf, conf.Slurm.Template)
}
