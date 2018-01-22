package pbs

import (
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// NewBackend returns a new PBS (Portable Batch System) HPCBackend instance.
func NewBackend(conf config.Config, reader tes.ReadOnlyServer, writer events.Writer) *compute.HPCBackend {
	b := &compute.HPCBackend{
		Name:      "pbs",
		SubmitCmd: "qsub",
		CancelCmd: "qdel",
		Conf:      conf,
		Template:  conf.PBS.Template,
		Event:     writer,
		Database:  reader,
		ExtractID: extractID,
	}
	return b
}

// extractID extracts the task id from the response returned by the `qsub` command.
// For PBS / Torque systems, `qsub` returns the task id
func extractID(in string) string {
	return in
}
