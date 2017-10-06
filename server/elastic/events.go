package elastic

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	elastic "gopkg.in/olivere/elastic.v5"
	"reflect"
	"strconv"
)

// Elastic provides an elasticsearch database server backend.
type Elastic struct {
	client *elastic.Client
	conf   config.Elastic
}

// NewElastic returns a new Elastic instance.
func NewElastic(conf config.Elastic) (*Elastic, error) {
	client, err := elastic.NewSimpleClient(
		elastic.SetURL(conf.URL),
	)
	if err != nil {
		return nil, err
	}
	return &Elastic{client, conf}, nil
}

func (es *Elastic) Counts() {
	agg := elastic.NewTermsAggregation().Field("state.keyword").Size(10).OrderByCountDesc()
	res, err := es.client.Search().
		Index(es.conf.Index).
		Type("task").
		Aggregation("state-counts", agg).
		Pretty(true).
		Do(context.Background())

	if err != nil {
		panic(err)
	}
	a, found := res.Aggregations.Terms("state-counts")
	if !found {
		panic("not found")
	}
	for _, b := range a.Buckets {
		k := b.Key.(string)
		i64, _ := strconv.ParseInt(k, 10, 32)
		i := int32(i64)
		fmt.Println(k, tes.State_name[i], b.DocCount)
	}
}

// Init initializing the Elasticsearch indices.
func (es *Elastic) Init(ctx context.Context) error {
	if exists, err := es.client.IndexExists(es.conf.Index).Do(ctx); err != nil {
		return err
	} else if !exists {
		if _, err := es.client.CreateIndex(es.conf.Index).Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

// CreateTask creates a new task.
func (es *Elastic) CreateTask(ctx context.Context, task *tes.Task) error {
	_, err := es.client.Update().
		Index(es.conf.Index).
		Type("task").
		Id(task.Id).
		// TODO need to be consistent with jsonpb
		Doc(task).
		DocAsUpsert(true).
		Do(ctx)
	return err
}

// ListTasks lists tasks, duh.
func (es *Elastic) ListTasks(ctx context.Context, req *tes.ListTasksRequest) ([]*tes.Task, error) {
	res, err := es.client.Search().
		Index(es.conf.Index).
		Type("task").
		Size(getPageSize(req)).
		// TODO sorting is broken
		Do(ctx)

	if err != nil {
		return nil, err
	}

	var tasks []*tes.Task
	var task tes.Task
	for _, item := range res.Each(reflect.TypeOf(task)) {
		i := item.(tes.Task)
		t := &i

		switch req.View {
		case tes.TaskView_BASIC:
			t = t.GetBasicView()
		case tes.TaskView_MINIMAL:
			t = t.GetMinimalView()
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
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
		Index(es.conf.Index).
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

// Write writes a task update event.
func (es *Elastic) Write(ev *events.Event) error {
	ctx := context.Background()

	switch ev.Type {
	case events.Type_TASK_CREATED:
		return es.CreateTask(ctx, ev.GetTask())

	case events.Type_SYSTEM_LOG:
		mar := jsonpb.Marshaler{}
		s, err := mar.MarshalToString(ev)
		if err != nil {
			return err
		}

		_, err = es.client.Index().
			Index(es.conf.Index).
			Type("task-syslog").
			BodyString(s).
			Do(ctx)
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

	_, err = es.client.Update().
		Index(es.conf.Index).
		Type("task").
		Id(task.Id).
		Doc(task).
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
