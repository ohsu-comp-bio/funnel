package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
)

// NoopRunner is useful during testing for creating a worker with a Runner
// that doesn't do anything.
type NoopRunner struct{}

// Run doesn't do anything, it's an empty function.
func (NoopRunner) Run(context.Context) {}

// NoopRunnerFactory returns a new NoopRunner.
func NoopRunnerFactory(c config.Worker, taskID string) Runner {
	return NoopRunner{}
}
