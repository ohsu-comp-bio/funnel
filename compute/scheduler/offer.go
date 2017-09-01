package scheduler

import (
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// Offer describes a node offered by a scheduler for a task.
// The Scores describe how well the task fits this node,
// which could be used by other a scheduler to pick the best offer.
type Offer struct {
	TaskID string
	Node   *pbs.Node
	Scores Scores
}

// NewOffer returns a new Offer instance.
func NewOffer(n *pbs.Node, t *tes.Task, s Scores) *Offer {
	return &Offer{
		TaskID: t.Id,
		Node:   n,
		Scores: s,
	}
}
