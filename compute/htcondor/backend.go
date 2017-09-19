package htcondor

import (
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
)

// NewBackend returns a new HtCondor HPCBackend instance.
func NewBackend(conf config.Config) *compute.HPCBackend {
	return compute.NewHPCBackend("htcondor", "condor_submit", conf, conf.Backends.HTCondor.Template)
}
