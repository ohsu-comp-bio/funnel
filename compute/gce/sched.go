package gce

// TODO
// - resource tracking via GCP APIs
// - provisioning limits, e.g. don't create more than 100 VMs, or
//   maybe use N VCPUs max, across all VMs
// - act on failed machines?
// - know how to shutdown machines

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

// NewBackend returns a new Google Cloud Engine Backend instance.
func NewBackend(conf config.Config, log *logger.Logger) (*Backend, error) {
	// TODO need GCE scheduler config validation. If zone is missing, nothing works.

	// Create a client for talking to the funnel scheduler
	client, err := scheduler.NewClient(conf.Scheduler)
	if err != nil {
		return nil, fmt.Errorf("can't connect scheduler client: %s", err)
	}

	// Create a client for talking to the GCE API
	gce, gerr := newClientFromConfig(conf)
	if gerr != nil {
		return nil, fmt.Errorf("can't connect GCE client: %s", gerr)
	}
	gce.log = log

	return &Backend{
		conf:   conf,
		client: client,
		gce:    gce,
		log:    log,
	}, nil
}

// Backend represents the GCE backend, which provides
// and interface for both scheduling and scaling.
type Backend struct {
	conf   config.Config
	client scheduler.Client
	gce    Client
	log    *logger.Logger
}

// GetOffer returns an offer based on available Google Cloud VM node instances.
func (s *Backend) GetOffer(j *tes.Task) *scheduler.Offer {
	offers := []*scheduler.Offer{}
	predicates := append(scheduler.DefaultPredicates, scheduler.NodeHasTag("gce"))

	for _, n := range s.getNodes() {
		// Filter out nodes that don't match the task request.
		// Checks CPU, RAM, disk space, etc.
		if !scheduler.Match(n, j, predicates) {
			continue
		}

		sc := scheduler.DefaultScores(n, j)
		/*
			    TODO?
			    if w.State == pbs.NodeState_Alive {
					  sc["startup time"] = 1.0
			    }
		*/
		weights := map[string]float32{}
		sc = sc.Weighted(weights)

		offer := scheduler.NewOffer(n, j, sc)
		offers = append(offers, offer)
	}

	// No matching nodes were found.
	if len(offers) == 0 {
		return nil
	}

	scheduler.SortByAverageScore(offers)
	return offers[0]
}

// getNodes returns a list of all GCE nodes and appends a set of
// uninitialized nodes, which the scheduler can use to create new node VMs.
func (s *Backend) getNodes() []*pbs.Node {

	// Get the nodes from the funnel server
	nodes := []*pbs.Node{}
	req := &pbs.ListNodesRequest{}
	resp, err := s.client.ListNodes(context.Background(), req)

	// If there's an error, return an empty list
	if err != nil {
		s.log.Error("failed GCE getNodes", err)
		return nodes
	}

	nodes = resp.Nodes

	// Include unprovisioned (template) nodes.
	// This is how the scheduler can schedule tasks to nodes that
	// haven't been started yet.
	for _, t := range s.gce.Templates() {
		t.Id = scheduler.GenNodeID("funnel")
		nodes = append(nodes, &t)
	}

	return nodes
}

// ShouldStartNode tells the scaler loop which nodes
// belong to this backend, basically.
func (s *Backend) ShouldStartNode(n *pbs.Node) bool {
	// Only start works that are uninitialized and have a gce template.
	tpl, ok := n.Metadata["gce-template"]
	return ok && tpl != "" && n.State == pbs.NodeState_UNINITIALIZED
}

// StartNode calls out to GCE APIs to start a new node instance.
func (s *Backend) StartNode(n *pbs.Node) error {

	// Get the template ID from the node metadata
	template, ok := n.Metadata["gce-template"]
	if !ok || template == "" {
		return fmt.Errorf("Could not get GCE template ID from metadata")
	}

	return s.gce.StartNode(template, s.conf.Server.RPCAddress(), n.Id)
}
