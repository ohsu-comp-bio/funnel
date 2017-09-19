package scheduler

import (
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// Backend is responsible for scheduling a task. It has a single method which
// is responsible for taking a Task and returning an Offer, or nil if there is
// no node matching the task request. An Offer includes the ID of the offered
// node.
//
// Offers include scores which describe how well the task fits the node.
// Scores may describe a wide variety of metrics: resource usage, packing,
// startup time, cost, etc. Scores and weights are used to control the behavior
// of schedulers, and to combine offers from multiple schedulers.
type Backend interface {
	// GetOffer returns an offer for a task.
	GetOffer(*tes.Task) *Offer
}

// Scaler represents a service that can start node instances, for example
// the Google Cloud Scheduler backend.
type Scaler interface {
	// StartNode is where the work is done to start a node instance,
	// for example, calling out to Google Cloud APIs.
	StartNode(*pbs.Node) error
	// ShouldStartNode allows scalers to filter out nodes they are interested in.
	// If "true" is returned, Scaler.StartNode() will be called with this Node.
	ShouldStartNode(*pbs.Node) bool
}
