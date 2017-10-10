package elastic

import "github.com/ohsu-comp-bio/funnel/logger"
import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	elastic "gopkg.in/olivere/elastic.v5"
	"reflect"
	"time"
)

var log = logger.Sub("elastic")

// QueueTask adds a task to the scheduler queue.
func (es *Elastic) QueueTask(task *tes.Task) error {
	return nil
}

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (es *Elastic) ReadQueue(n int) []*tes.Task {
	ctx := context.Background()

	q := elastic.NewTermQuery("state", tes.State_QUEUED)
	res, err := es.client.Search().
		Index(es.taskIndex).
		// TODO
		//Sort("id", true).
		Query(q).
		Do(ctx)

	if err != nil {
		fmt.Println(err)
		return nil
	}

	var tasks []*tes.Task
	var task tes.Task

	for i, item := range res.Each(reflect.TypeOf(task)) {
		t := item.(tes.Task)
		tasks = append(tasks, &t)
		if i == n {
			break
		}
	}

	return tasks
}

// GetNode gets a node
func (es *Elastic) GetNode(ctx context.Context, req *pbs.GetNodeRequest) (*pbs.Node, error) {
	res, err := es.client.Get().
		Index(es.nodeIndex).
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
	return node, nil
}

// PutNode puts a node in the database.
//
// For optimisic locking, if the node already exists and node.Version
// doesn't match the version in the database, an error is returned.
func (es *Elastic) PutNode(ctx context.Context, node *pbs.Node) (*pbs.PutNodeResponse, error) {
	res, err := es.client.Get().
		Index(es.nodeIndex).
		Id(node.Id).
		Do(ctx)

	existing := &pbs.Node{}
	if err == nil {
		jsonpb.Unmarshal(bytes.NewReader(*res.Source), existing)
	}

	if existing.GetVersion() != 0 && node.Version != existing.GetVersion() {
		return nil, fmt.Errorf("Version outdated")
	}
	node.Version = time.Now().UnixNano()

	mar := jsonpb.Marshaler{}
	s, err := mar.MarshalToString(node)
	if err != nil {
		return nil, err
	}

	_, err = es.client.Index().
		Index(es.nodeIndex).
		Id(node.Id).
		BodyString(s).
		Do(ctx)
	return &pbs.PutNodeResponse{}, err
}

// DeleteNode deletes a node by ID.
func (es *Elastic) DeleteNode(ctx context.Context, node *pbs.Node) error {
	_, err := es.client.Delete().
		Index(es.nodeIndex).
		Id(node.Id).
		Do(ctx)
	return err
}

// ListNodes is an API endpoint that returns a list of nodes.
func (es *Elastic) ListNodes(ctx context.Context, req *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error) {
	res, err := es.client.Search().
		Index(es.nodeIndex).
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
		resp.Nodes = append(resp.Nodes, node)
	}

	return resp, nil
}
