package slurm

import (
	"regexp"

	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// NewBackend returns a new Slurm HPCBackend instance.
func NewBackend(conf config.Config, reader tes.ReadOnlyServer, writer events.Writer) *compute.HPCBackend {
	b := &compute.HPCBackend{
		Name:      "slurm",
		SubmitCmd: "sbatch",
		CancelCmd: "scancel",
		Conf:      conf,
		Template:  conf.Slurm.Template,
		Event:     writer,
		Database:  reader,
		ExtractID: extractID,
	}
	return b
}

// extractID extracts the task id from the response returned by the `sbatch` command.
// Example response:
// Submitted batch job 2
func extractID(in string) string {
	re := regexp.MustCompile("(Submitted batch job )([0-9]+)\n$")
	return re.ReplaceAllString(in, "$2")
}
