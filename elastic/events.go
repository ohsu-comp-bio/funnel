package elastic

import (
	"bytes"
	"context"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	elastic "gopkg.in/olivere/elastic.v5"
	"reflect"
)

type Elastic struct {
	client *elastic.Client
	conf   config.Elastic
}

func NewElastic(conf config.Elastic) (*Elastic, error) {
  // TODO simple client doesn't work for clusters
	client, err := elastic.NewSimpleClient(
		elastic.SetURL(conf.URL),
  )
	if err != nil {
		return nil, err
	}
	return &Elastic{client, conf}, nil
}

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

func (es *Elastic) CreateTask(ctx context.Context, task *tes.Task) error {
	_, err := es.client.Update().
		Index(es.conf.Index).
		Type("task").
		Id(task.Id).
		Doc(task).
		DocAsUpsert(true).
		Do(ctx)
	return err
}

func (es *Elastic) ListTasks(ctx context.Context) ([]*tes.Task, error) {
	res, err := es.client.Search().
		Index(es.conf.Index).
		Type("task").
    // TODO
    Size(1000).
    // TODO sorting is broken
		Do(ctx)

	if err != nil {
		return nil, err
	}

	var tasks []*tes.Task
	var task tes.Task
	for _, item := range res.Each(reflect.TypeOf(task)) {
		t := item.(tes.Task)
		tasks = append(tasks, &t)
	}

	return tasks, nil
}

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

func (es *Elastic) Write(ev *events.Event) error {
	ctx := context.Background()

	if ev.Type == events.Type_TASK_CREATED {
		return es.CreateTask(ctx, ev.GetTask())
	}

	task, err := es.GetTask(ctx, ev.Id)
	if err != nil {
		return err
	}

	err = events.TaskBuilder{task}.Write(ev)
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

func tail(s string, sizeBytes int) string {
	b := []byte(s)
	if len(b) > sizeBytes {
		return string(b[:sizeBytes])
	}
	return string(b)
}

/*
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
