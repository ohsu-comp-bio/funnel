package elastic

import (
	"bytes"

	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	elastic "gopkg.in/olivere/elastic.v5"
)

// GetTask gets a task by ID.
func (es *Elastic) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	g := es.client.Get().
		Index(es.taskIndex).
		Type("task").
		Id(req.Id)

	switch req.View {
	case tes.TaskView_BASIC:
		g = g.FetchSource(true).FetchSourceContext(basic)
	case tes.TaskView_MINIMAL:
		g = g.FetchSource(true).FetchSourceContext(minimal)
	}

	res, err := g.Do(ctx)

	if elastic.IsNotFound(err) {
		return nil, tes.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	task := &tes.Task{}
	err = jsonpb.Unmarshal(bytes.NewReader(*res.Source), task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// ListTasks lists tasks, duh.
func (es *Elastic) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {

	pageSize := tes.GetPageSize(req.GetPageSize())
	q := es.client.Search().
		Index(es.taskIndex).
		Type("task")

	if req.PageToken != "" {
		q = q.SearchAfter(req.PageToken)
	}

	if req.StateFilter != tes.Unknown {
		q = q.Query(elastic.NewTermQuery("state", req.StateFilter.String()))
	}

	q = q.Sort("id", false).Size(pageSize)

	switch req.View {
	case tes.TaskView_BASIC:
		q = q.FetchSource(true).FetchSourceContext(basic)
	case tes.TaskView_MINIMAL:
		q = q.FetchSource(true).FetchSourceContext(minimal)
	}

	res, err := q.Do(ctx)
	if err != nil {
		return nil, err
	}

	resp := &tes.ListTasksResponse{}
	for i, hit := range res.Hits.Hits {
		t := &tes.Task{}
		err := jsonpb.Unmarshal(bytes.NewReader(*hit.Source), t)
		if err != nil {
			return nil, err
		}

		if i == pageSize-1 {
			resp.NextPageToken = t.Id
		}

		resp.Tasks = append(resp.Tasks, t)
	}

	return resp, nil
}
