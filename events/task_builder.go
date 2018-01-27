package events

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// TaskBuilder aggregates events into an in-memory Task object.
type TaskBuilder struct {
	*tes.Task
}

// WriteEvent updates the Task object.
func (tb TaskBuilder) WriteEvent(ctx context.Context, ev *Event) error {
	t := tb.Task
	t.Id = ev.Id

	switch ev.Type {
	case Type_STATE:
		to := ev.GetState()
		if err := tes.ValidateTransition(t.GetState(), ev.GetState()); err != nil {
			return err
		}
		t.State = to

	case Type_START_TIME:
		t.StartTime = ev.GetStartTime()

	case Type_END_TIME:
		t.EndTime = ev.GetEndTime()

	case Type_OUTPUTS:
		t.OutputsLog = ev.GetOutputs().Value

	case Type_METADATA:
		t.Metadata = ev.GetMetadata().Value

	case Type_EXIT_CODE:
		t.ExitCode = ev.GetExitCode()

	case Type_STDOUT:
		t.Stdout += ev.GetStdout()

	case Type_STDERR:
		t.Stderr += ev.GetStderr()
	}

	return nil
}
