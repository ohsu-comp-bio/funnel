package htcondor

import (
	"regexp"

	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// NewBackend returns a new HTCondor backend instance.
func NewBackend(conf config.Config, reader tes.ReadOnlyServer, writer events.Writer) *compute.HPCBackend {
	b := &compute.HPCBackend{
		Name:      "htcondor",
		SubmitCmd: "condor_submit",
		CancelCmd: "condor_rm",
		Conf:      conf,
		Template:  conf.HTCondor.Template,
		Event:     writer,
		Database:  reader,
		ExtractID: extractID,
	}
	return b
}

// extractID extracts the task id from the response returned by the `condor_submit` command.
// Example response:
// Submitting job(s).
// 1 job(s) submitted to cluster 1.
func extractID(in string) string {
	re := regexp.MustCompile("(.*\\.\n.*)([0-9]+)\\.\n")
	return re.ReplaceAllString(in, "$2")
}
