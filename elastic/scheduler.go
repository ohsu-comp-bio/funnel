package elastic

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/events"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	elastic "gopkg.in/olivere/elastic.v5"
	"reflect"
)

// QueueTask adds a task to the scheduler queue.
func (es *Elastic) QueueTask(task *tes.Task) error {
	return nil
}

// ReadQueue returns a slice of queued Tasks. Up to "n" tasks are returned.
func (es *Elastic) ReadQueue(n int) []*tes.Task {
	ctx := context.Background()

	q := elastic.NewTermQuery("state", tes.State_QUEUED)
	res, err := es.client.Search().
		Index(es.conf.Index).
		Type("task").
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

// AssignTask assigns a task to a node. This updates the task state to Initializing,
// and updates the node (calls UpdateNode()).
func (es *Elastic) AssignTask(t *tes.Task, w *pbs.Node) error {
	err := es.Write(events.NewState(t.Id, 0, tes.State_INITIALIZING))
	if err != nil {
		return err
	}

	ctx := context.Background()
	node, err := es.GetNode(ctx, &pbs.GetNodeRequest{
		Id: w.Id,
	})
	if err != nil {
		return err
	}

	node.TaskIds = append(node.TaskIds, t.Id)
	return es.PutNode(ctx, node)
}

// GetNode gets a node
func (es *Elastic) GetNode(ctx context.Context, req *pbs.GetNodeRequest) (*pbs.Node, error) {
	return nil, nil
}

// PutNode
func (es *Elastic) PutNode(ctx context.Context, node *pbs.Node) error {
	return nil
}

func (es *Elastic) DeleteNode(ctx context.Context, id string) error {
	return nil
}

// CheckNodes is used by the scheduler to check for dead/gone nodes.
// This is not an RPC endpoint
func (es *Elastic) CheckNodes() error {
	return nil
}

// ListNodes is an API endpoint that returns a list of nodes.
func (es *Elastic) ListNodes(ctx context.Context, req *pbs.ListNodesRequest) (*pbs.ListNodesResponse, error) {
	return nil, nil
}
