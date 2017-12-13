package pbs

import (
	"context"
	"encoding/xml"
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"os/exec"
)

// NewBackend returns a new PBS (Portable Batch System) HPCBackend instance.
func NewBackend(ctx context.Context, conf config.Config, reader tes.ReadOnlyServer, writer events.Writer) *compute.HPCBackend {
	b := &compute.HPCBackend{
		Name:          "pbs",
		SubmitCmd:     "qsub",
		CancelCmd:     "qdel",
		Conf:          conf,
		Template:      conf.PBS.Template,
		Event:         writer,
		Database:      reader,
		ExtractID:     extractID,
		MapStates:     mapStates,
		ReconcileRate: conf.GridEngine.ReconcileRate,
	}
	go b.Reconcile(ctx)
	return b
}

func extractID(in string) string {
	return in
}

type job struct {
	JobID      string `xml:"Job_Id"`
	JobState   string `xml:"job_state"`
	ExitStatus int    `xml:"exit_status"`
}

type xmlRecord struct {
	XMLName xml.Name `xml:"Data"`
	Job     []job
}

func mapStates(ids []string) ([]*compute.HPCTaskState, error) {
	var output []*compute.HPCTaskState

	idSet := make(map[string]interface{})
	for _, i := range ids {
		idSet[i] = nil
	}

	cmd := exec.Command("qstat", "-x")
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	res := xmlRecord{}
	err = xml.Unmarshal(stdout, &res)
	if err != nil {
		return nil, err
	}

	for _, j := range res.Job {
		if _, ok := idSet[j.JobID]; !ok {
			continue
		}
		state := stateMap[j.JobState]
		switch state {
		case "Queued":
			output = append(output, &compute.HPCTaskState{ID: j.JobID, TESState: tes.Queued, State: state})
		case "Running":
			output = append(output, &compute.HPCTaskState{ID: j.JobID, TESState: tes.Running, State: state})
		case "Moving":
			output = append(output, &compute.HPCTaskState{ID: j.JobID, TESState: tes.Running, State: state})
		case "Waiting":
			// noop
		case "Suspended":
			// noop
		case "Exiting":
			// noop:
		case "Complete":
			if j.ExitStatus == 0 {
				output = append(output, &compute.HPCTaskState{ID: j.JobID, TESState: tes.Complete, State: state})
			} else {
				output = append(output, &compute.HPCTaskState{ID: j.JobID, TESState: tes.SystemError, State: state, Reason: "exited with non-zero status"})
			}
		}
	}
	return output, nil
}

var stateMap = map[string]string{
	"C": "Complete",
	"E": "Exiting",
	"H": "Held",
	"Q": "Queued",
	"R": "Running",
	"S": "Suspended",
	"T": "Moving",
	"W": "Waiting",
}
