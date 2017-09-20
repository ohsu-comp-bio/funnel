package server

import elastic "gopkg.in/olivere/elastic.v5"
import (
  "context"
	proto "github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

type Config struct {
  MaxLogSize int
}

type Elastic struct {
  client *elastic.Client
  conf Config
}

func NewElastic(url string) (*Elastic, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(url),
	)
  if err != nil {
    return nil, err
  }
  conf := Config{MaxLogSize: 1000}
  return &Elastic{client, conf}, nil
}

func (es *Elastic) Init(ctx context.Context) error {
	if exists, err := client.IndexExists("funnel").Do(ctx); err != nil {
    return err
	} else if !exists {
		if _, err := client.CreateIndex("funnel").Do(ctx); err != nil {
      return err
		}
	}
  return nil
}

func (es *Elastic) GetTask(ctx context.Context, id string) error {
  get1, err := es.client.Get().
    Index("funnel").
    Type("task").
    Id("1").
    Do(ctx)

  if err != nil {
    return err
  }

if get1.Found {
    fmt.Printf("Got document %s in version %d from index %s, type %s\n", get1.Id, get1.Version, get1.Index, get1.Type)
}
}

func (es *Elastic) CreateTask(ctx context.Context, task *tes.Task) error {
}

// CreateEvent creates an event for the server to handle.
func (es *Elastic) CreateEvent(ctx context.Context, ev *events.Event) (*events.CreateEventResponse, error) {
	var err error

	tl := &tes.TaskLog{}
	el := &tes.ExecutorLog{}

	switch req.Type {
	case events.Type_TASK_STATE:
    t.State = ev.GetState()

	case events.Type_TASK_START_TIME:
		tl.StartTime = ptypes.TimestampString(req.GetStartTime())

	case events.Type_TASK_END_TIME:
		tl.EndTime = ptypes.TimestampString(req.GetEndTime())

	case events.Type_TASK_OUTPUTS:
		tl.Outputs = req.GetOutputs().Value

	case events.Type_TASK_METADATA:
		tl.Metadata = req.GetMetadata().Value

	case events.Type_EXECUTOR_START_TIME:
		el.StartTime = ptypes.TimestampString(req.GetStartTime())

	case events.Type_EXECUTOR_END_TIME:
		el.EndTime = ptypes.TimestampString(req.GetEndTime())

	case events.Type_EXECUTOR_EXIT_CODE:
		el.ExitCode = req.GetExitCode()

	case events.Type_EXECUTOR_HOST_IP:
		el.HostIp = req.GetHostIp()

	case events.Type_EXECUTOR_PORTS:
		el.Ports = req.GetPorts().Value

	case events.Type_EXECUTOR_STDOUT:
		el.Stdout = req.GetStdout()

	case events.Type_EXECUTOR_STDERR:
		el.Stderr = req.GetStderr()
	}

	if err != nil {
		return nil, err
	}

	return &events.CreateEventResponse{}, nil
}

func tail(s string, sizeBytes int) string {
  b := []byte(s)
  if len(b) > size {
    return b[:size]
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
