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
	client    *elastic.Client
	conf      config.Elastic
	taskIndex string
	nodeIndex string
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
	}, nil
}

func (es *Elastic) Close() error {
  es.client.Stop()
  return nil
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
          },
          "state": {
            "type": "keyword"
          },
          "inputs": {
            "type": "nested"
          },
          "logs": {
            "type": "nested",
            "properties": {
              "logs": {
                "type": "nested"
              }
            }
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

	pageSize := tes.GetPageSize(req.GetPageSize())
	q := es.client.Search().
		Index(es.taskIndex).
		Type("task")

	if req.PageToken != "" {
		q = q.SearchAfter(req.PageToken)
	}

	q = q.SortBy(idFieldSort).Size(pageSize)

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

var minimal = elastic.NewFetchSourceContext(true).Include("id", "state")
var basic = elastic.NewFetchSourceContext(true).
	Exclude("logs.logs.stderr", "logs.logs.stdout", "inputs.contents")

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
	return es.WriteContext(context.Background(), ev)
}

var updateTaskLogs = `
if (ctx._source.logs == null) {
  ctx._source.logs = new ArrayList();
}

// Ensure the task logs array is long enough.
for (; params.attempt > ctx._source.logs.length - 1; ) {
  Map m = new HashMap();
  m.logs = new ArrayList();
  ctx._source.logs.add(m);
}

// Set the field.
ctx._source.logs[params.attempt][params.field] = params.value;
`

var updateExecutorLogs = `
if (ctx._source.logs == null) {
  ctx._source.logs = new ArrayList();
}

// Ensure the task logs array is long enough.
for (; params.attempt > ctx._source.logs.length - 1; ) {
  Map m = new HashMap();
  m.logs = new ArrayList();
  ctx._source.logs.add(m);
}

// Ensure the executor logs array is long enough.
for (; params.index > ctx._source.logs[params.attempt].logs.length - 1; ) {
  Map m = new HashMap();
  ctx._source.logs[params.attempt].logs.add(m);
}

// Set the field.
ctx._source.logs[params.attempt].logs[params.index][params.field] = params.value;
`

func taskLogUpdate(attempt uint32, field string, value interface{}) *elastic.Script {
	return elastic.NewScript(updateTaskLogs).
		Lang("painless").
		Param("attempt", attempt).
		Param("field", field).
		Param("value", value)
}

func execLogUpdate(attempt, index uint32, field string, value interface{}) *elastic.Script {
	return elastic.NewScript(updateExecutorLogs).
		Lang("painless").
		Param("attempt", attempt).
		Param("index", index).
		Param("field", field).
		Param("value", value)
}

// WriteContext writes a task update event.
func (es *Elastic) WriteContext(ctx context.Context, ev *events.Event) error {
	// Skipping system logs for now. Will add them to the task logs when this PR is resolved (soon):
	// https://github.com/ga4gh/task-execution-schemas/pull/80
	if ev.Type == events.Type_SYSTEM_LOG {
		return nil
	}

	u := es.client.Update().
		Index(es.taskIndex).
		Type("task").
		RetryOnConflict(3).
		Id(ev.Id)

	switch ev.Type {
	case events.Type_TASK_STATE:
		u = u.Doc(map[string]string{"state": ev.GetState().String()})

	case events.Type_TASK_START_TIME:
		u = u.Script(taskLogUpdate(ev.Attempt, "start_time", events.TimestampString(ev.GetStartTime())))

	case events.Type_TASK_END_TIME:
		u = u.Script(taskLogUpdate(ev.Attempt, "end_time", events.TimestampString(ev.GetEndTime())))

	case events.Type_TASK_OUTPUTS:
		u = u.Script(taskLogUpdate(ev.Attempt, "outputs", ev.GetOutputs().Value))

	case events.Type_TASK_METADATA:
		u = u.Script(taskLogUpdate(ev.Attempt, "metadata", ev.GetMetadata().Value))

	case events.Type_EXECUTOR_START_TIME:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "start_time", events.TimestampString(ev.GetStartTime())))

	case events.Type_EXECUTOR_END_TIME:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "end_time", events.TimestampString(ev.GetEndTime())))

	case events.Type_EXECUTOR_EXIT_CODE:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "exit_code", ev.GetExitCode()))

	case events.Type_EXECUTOR_HOST_IP:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "host_ip", ev.GetHostIp()))

	case events.Type_EXECUTOR_PORTS:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "ports", ev.GetPorts().Value))

	case events.Type_EXECUTOR_STDOUT:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "stdout", ev.GetStdout()))

	case events.Type_EXECUTOR_STDERR:
		u = u.Script(execLogUpdate(ev.Attempt, ev.Index, "stderr", ev.GetStderr()))
	}

	_, err := u.Do(ctx)
	return err
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
