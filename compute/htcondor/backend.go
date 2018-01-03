package htcondor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/compute"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"os/exec"
	"regexp"
	"strings"
)

// NewBackend returns a new HTCondor backend instance.
func NewBackend(ctx context.Context, conf config.Config, reader tes.ReadOnlyServer, writer events.Writer) *compute.HPCBackend {
	b := &compute.HPCBackend{
		Name:          "htcondor",
		SubmitCmd:     "condor_submit",
		CancelCmd:     "condor_rm",
		Conf:          conf,
		Template:      conf.HTCondor.Template,
		Event:         writer,
		Database:      reader,
		ExtractID:     extractID,
		MapStates:     mapStates,
		ReconcileRate: conf.HTCondor.ReconcileRate,
	}
	go b.Reconcile(ctx)
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

type record struct {
	ClusterId int //nolint
	JobStatus int
	ExitCode  int
}

func mapStates(ids []string) ([]*compute.HPCTaskState, error) {
	var output []*compute.HPCTaskState

	qcmd := exec.Command("condor_q", "-json", "-attributes", "ClusterId,JobStatus,ExitCode", strings.Join(ids, ","))
	qout, err := qcmd.Output()
	if err != nil {
		return nil, err
	}

	qparsed := []record{}
	err = json.Unmarshal(qout, &qparsed)
	if err != nil {
		return nil, err
	}

	hcmd := exec.Command("condor_history", "-json", "-attributes", "ClusterId,JobStatus,ExitCode", strings.Join(ids, ","))
	hout, err := hcmd.Output()
	if err != nil {
		return nil, err
	}

	hparsed := []record{}
	err = json.Unmarshal(hout, &hparsed)
	if err != nil {
		return nil, err
	}

	parsed := append(qparsed, hparsed...)

	for _, t := range parsed {
		id := fmt.Sprintf("%v", t.ClusterId)
		exitcode := t.ExitCode
		state := stateMap[t.JobStatus]

		switch state {
		case "Idle":
			err = checkIdleStatus(id)
			if err != nil {
				output = append(output, &compute.HPCTaskState{
					ID: id, TESState: tes.SystemError, State: state, Reason: err.Error(), Remove: true,
				})
			} else {
				output = append(output, &compute.HPCTaskState{ID: id, TESState: tes.Queued, State: state})
			}

		case "Running":
			output = append(output, &compute.HPCTaskState{ID: id, TESState: tes.Running, State: state})

		case "Held":
			output = append(output, &compute.HPCTaskState{ID: id, TESState: tes.Running, State: state})

		case "Removed":
			output = append(output, &compute.HPCTaskState{ID: id, TESState: tes.Canceled, State: state, Reason: "task was canceled"})

		case "Submission_err":
			output = append(output, &compute.HPCTaskState{ID: id, TESState: tes.SystemError, State: state, Reason: "task encountered submission error"})

		case "Completed":
			if exitcode == 0 {
				output = append(output, &compute.HPCTaskState{ID: id, TESState: tes.Complete, State: state})
			} else {
				output = append(output, &compute.HPCTaskState{ID: id, TESState: tes.SystemError, State: state, Reason: "Funnel worker exited with non-zero status"})
			}
		}
	}
	return output, nil
}

func checkIdleStatus(id string) error {
	cmd := exec.Command("condor_q", "-analyze", id)
	stdout, _ := cmd.Output()
	msg := "No machines matched the jobs's constraints"
	if stdout != nil && strings.Contains(string(stdout), msg) {
		return fmt.Errorf(msg)
	}
	return nil
}

var stateMap = map[int]string{
	1: "Idle",
	2: "Running",
	3: "Removed",
	4: "Completed",
	5: "Held",
	6: "Submission_err",
}
