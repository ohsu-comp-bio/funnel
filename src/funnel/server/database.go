package server

import (
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
)

// Database represents the interface to the database used by the scheduler, scaler, etc.
// Mostly, this exists so it can be mocked during testing.
type Database interface {
	tes.TaskServiceServer
	pbf.SchedulerServer
}
