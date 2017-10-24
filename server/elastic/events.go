package elastic

import (
	"bytes"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	elastic "gopkg.in/olivere/elastic.v5"
	"time"
)

var idFieldSort = elastic.NewFieldSort("id").
	Desc().
	// Handles the case where there are no documents in the index.
	UnmappedType("keyword")

// Elastic provides an elasticsearch database server backend.
type Elastic struct {
	client      *elastic.Client
	conf        config.Elastic
	taskIndex   string
	nodeIndex   string
	eventsIndex string
}

// NewElastic returns a new Elastic instance.
func NewElastic(conf config.Elastic) (*Elastic, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(conf.URL),
		elastic.SetSniff(false),
		elastic.SetRetrier(
			elastic.NewBackoffRetrier(
				elastic.NewExponentialBackoff(time.Millisecond*50, time.Minute),
			),
		),
	)
	if err != nil {
		return nil, err
	}
	return &Elastic{
		client,
		conf,
		conf.IndexPrefix + "-tasks",
		conf.IndexPrefix + "-nodes",
		conf.IndexPrefix + "-events",
	}, nil
}

func (es *Elastic) initIndex(ctx context.Context, name, body string) error {
	exists, err := es.client.
		IndexExists(name).
		Do(ctx)

	if err != nil {
		return err
	} else if !exists {
		if _, err := es.client.CreateIndex(name).Body(body).Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Init initializing the Elasticsearch indices.
func (es *Elastic) Init(ctx context.Context) error {
	taskMappings := `{
    "mappings": {
      "task":{
        "properties":{
          "id": {
            "type": "keyword"
          }
        }
      }
    }
  }`
	if err := es.initIndex(ctx, es.taskIndex, taskMappings); err != nil {
		return err
	}
	if err := es.initIndex(ctx, es.nodeIndex, ""); err != nil {
		return err
	}
	if err := es.initIndex(ctx, es.eventsIndex, ""); err != nil {
		return err
	}
	return nil
}

// CreateTask creates a new task.
func (es *Elastic) CreateTask(ctx context.Context, task *tes.Task) error {
	mar := jsonpb.Marshaler{}
	s, err := mar.MarshalToString(task)
	if err != nil {
		return err
	}

	_, err = es.client.Index().
		Index(es.taskIndex).
		Type("task").
		Id(task.Id).
		BodyString(s).
		Do(ctx)
	return err
}

// ListTasks lists tasks, duh.
func (es *Elastic) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {

	pageSize := getPageSize(req)
	q := es.client.Search().
		Index(es.taskIndex).
		Type("task")

	if req.PageToken != "" {
		q = q.SearchAfter(req.PageToken)
	}

	q = q.SortBy(idFieldSort).Size(pageSize)

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

		switch req.View {
		case tes.TaskView_BASIC:
			t = t.GetBasicView()
		case tes.TaskView_MINIMAL:
			t = t.GetMinimalView()
		}

		resp.Tasks = append(resp.Tasks, t)
	}

	return resp, nil
}

func getPageSize(req *tes.ListTasksRequest) int {
	pageSize := 256

	if req.PageSize != 0 {
		pageSize = int(req.GetPageSize())
		if pageSize > 2048 {
			pageSize = 2048
		}
		if pageSize < 50 {
			pageSize = 50
		}
	}
	return pageSize
}

// GetTask gets a task by ID.
func (es *Elastic) GetTask(ctx context.Context, id string) (*tes.Task, error) {
	res, err := es.client.Get().
		Index(es.taskIndex).
		Type("task").
		Id(id).
		Do(ctx)

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

func (es *Elastic) updateTask(ctx context.Context, task *tes.Task) error {
	mar := jsonpb.Marshaler{}
	s, err := mar.MarshalToString(task)

	if err != nil {
		return err
	}

	_, err = es.client.Index().
		Index(es.taskIndex).
		Type("task").
		Id(task.Id).
		BodyString(s).
		Do(ctx)
	return err
}

/*
func tail(s string, sizeBytes int) string {
	b := []byte(s)
	if len(b) > sizeBytes {
		return string(b[:sizeBytes])
	}
	return string(b)
}

func updateExecutorLogs(tx *bolt.Tx, id string, el *tes.ExecutorLog) error {
	// Check if there is an existing task log
	o := tx.Bucket(ExecutorLogs).Get([]byte(id))
	if o != nil {
		// There is an existing log in the DB, load it
		existing := &tes.ExecutorLog{}

    el.Stdout = tail(existing.Stdout + el.Stdout, es.conf.MaxLogSize)
    el.Stderr = tail(existing.Stderr + el.Stderr, es.conf.MaxLogSize)

		// Merge the updates into the existing.
		proto.Merge(existing, el)
		// existing is updated, so set that to el which will get saved below.
		el = existing
	}
}
*/

// Write writes a task update event.
func (es *Elastic) Write(ctx context.Context, ev *events.Event) error {
	mar := jsonpb.Marshaler{}
	s, err := mar.MarshalToString(ev)

	_, err = es.client.Index().
		Index(es.eventsIndex).
		Type("event").
		BodyString(s).
		Do(ctx)

	if err != nil {
		return err
	}

	task, err := es.GetTask(ctx, ev.Id)
	if err != nil {
		return err
	}

	err = events.TaskBuilder{Task: task}.Write(ev)
	if err != nil {
		return err
	}

	return es.updateTask(ctx, task)
}
