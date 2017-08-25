package server

import (
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	tl "github.com/ohsu-comp-bio/funnel/proto/tasklogger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// Database represents the interface to the database used by the scheduler, scaler, etc.
// Mostly, this exists so it can be mocked during testing.
type Database interface {
	tes.TaskServiceServer
	tl.TaskLoggerServiceServer
	pbs.SchedulerServiceServer
}
