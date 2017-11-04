package elastic

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	elastic "gopkg.in/olivere/elastic.v5"
)

// QueueTask adds a task to the scheduler queue.
func (es *Elastic) QueueTask(task *tes.Task) error {
	return nil
}

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (es *Elastic) ReadQueue(n int) []*tes.Task {
	ctx := context.Background()

	q := elastic.NewTermQuery("state.keyword", tes.State_QUEUED.String())
	res, err := es.client.Search().
		Index(es.taskIndex).
		Type("task").
		Size(n).
		Sort("id", true).
		Query(q).
		Do(ctx)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	var tasks []*tes.Task
	for _, hit := range res.Hits.Hits {
		t := &tes.Task{}
		err := jsonpb.Unmarshal(bytes.NewReader(*hit.Source), t)
		if err != nil {
			continue
		}

		t = t.GetBasicView()
		tasks = append(tasks, t)
	}

	return tasks
}

// GetNode gets a node
func (es *Elastic) GetNode(ctx context.Context, req *pbs.GetNodeRequest) (*pbs.Node, error) {
	res, err := es.client.Get().
		Index(es.nodeIndex).
		Type("node").
		Id(req.Id).
		Do(ctx)

	if elastic.IsNotFound(err) {
		return nil, grpc.Errorf(codes.NotFound, fmt.Sprintf("%v: nodeID: %s", err.Error(), req.Id))
	}
	if err != nil {
		return nil, err
	}

	node := &pbs.Node{}
	err = jsonpb.Unmarshal(bytes.NewReader(*res.Source), node)
	if err != nil {
		return nil, err
	}
	// Must happen after the unmarshal
	node.Version = *res.Version
	return node, nil
}

// PutNode puts a node in the database.
//
// For optimisic locking, if the node already exists and node.Version
// doesn't match the version in the database, an error is returned.
func (es *Elastic) PutNode(ctx context.Context, node *pbs.Node) (*pbs.PutNodeResponse, error) {
	g := es.client.Get().
		Index(es.nodeIndex).
		Type("node").
		Preference("_primary").
		Id(node.Id)

		// If the version is 0, then this should be creating a new node.
	if node.GetVersion() != 0 {
		g = g.Version(node.GetVersion())
	}

	res, err := g.Do(ctx)

	if err != nil && !elastic.IsNotFound(err) {
		return nil, err
	}

	existing := &pbs.Node{}
	if err == nil {
		jsonpb.Unmarshal(bytes.NewReader(*res.Source), existing)
	}

	err = scheduler.UpdateNode(ctx, &TES{Elastic: es}, node, existing)
	if err != nil {
		return nil, err
	}

	mar := jsonpb.Marshaler{}
	s, err := mar.MarshalToString(node)
	if err != nil {
		return nil, err
	}

	i := es.client.Index().
		Index(es.nodeIndex).
		Type("node").
		Id(node.Id).
		Refresh("true").
		BodyString(s)

	if node.GetVersion() != 0 {
		i = i.Version(node.GetVersion())
	}
	_, err = i.Do(ctx)

	return &pbs.PutNodeResponse{}, err
}

// DeleteNode deletes a node by ID.
func (es *Elastic) DeleteNode(ctx context.Context, node *pbs.Node) error {
	_, err := es.client.Delete().
		Index(es.nodeIndex).
		Type("node").
		Id(node.Id).
		Version(node.Version).
		Refresh("true").
		Do(ctx)
	return err
}

// ListNodes is an API endpoint that returns a list of nodes.
func (es *Elastic) ListNodes(ctx context.Context, req *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error) {
	res, err := es.client.Search().
		Index(es.nodeIndex).
		Type("node").
		Version(true).
		Size(1000).
		Do(ctx)

	if err != nil {
		return nil, err
	}

	resp := &pbs.ListNodesResponse{}
	for _, hit := range res.Hits.Hits {
		node := &pbs.Node{}
		err = jsonpb.Unmarshal(bytes.NewReader(*hit.Source), node)
		if err != nil {
			return nil, err
		}
		node.Version = *hit.Version
		resp.Nodes = append(resp.Nodes, node)
	}

	return resp, nil
}
