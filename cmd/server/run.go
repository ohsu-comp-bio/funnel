package server

import (
	"context"
	"fmt"
	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/compute/batch"
	"github.com/ohsu-comp-bio/funnel/compute/gridengine"
	"github.com/ohsu-comp-bio/funnel/compute/htcondor"
	"github.com/ohsu-comp-bio/funnel/compute/local"
	"github.com/ohsu-comp-bio/funnel/compute/noop"
	"github.com/ohsu-comp-bio/funnel/compute/pbs"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/compute/slurm"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	schedProto "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/server"
	"github.com/ohsu-comp-bio/funnel/server/boltdb"
	"github.com/ohsu-comp-bio/funnel/server/datastore"
	"github.com/ohsu-comp-bio/funnel/server/dynamodb"
	"github.com/ohsu-comp-bio/funnel/server/elastic"
	"github.com/ohsu-comp-bio/funnel/server/mongodb"
	"strings"
)

// Run runs the "server run" command.
func Run(ctx context.Context, conf config.Config) error {
	log := logger.NewLogger("server", conf.Logger)
	s, err := NewServer(conf, log)
	if err != nil {
		return err
	}
	return s.Run(ctx)
}

// Server is a Funnel server + scheduler.
type Server struct {
	*server.Server
	*scheduler.Scheduler
}

// NewServer returns a new Funnel server + scheduler based on the given config.
func NewServer(conf config.Config, log *logger.Logger) (*Server, error) {
	if log == nil {
		log = logger.NewLogger("server", conf.Logger)
	}
	log.Debug("NewServer", "config", conf)

	ctx := context.Background()
	var reader tes.ReadOnlyServer
	var nodes schedProto.SchedulerServiceServer
	var sched *scheduler.Scheduler
	var queue scheduler.TaskQueue

	writers := events.MultiWriter{}

	switch strings.ToLower(conf.Database) {
	case "boltdb":
		b, err := boltdb.NewBoltDB(conf.BoltDB)
		if err != nil {
			return nil, dberr(err)
		}
		reader = b
		nodes = b
		queue = b
		writers.Append(b)

	case "datastore":
		d, err := datastore.NewDatastore(conf.Datastore)
		if err != nil {
			return nil, dberr(err)
		}
		reader = d
		writers.Append(d)

	case "dynamodb":
		d, err := dynamodb.NewDynamoDB(conf.DynamoDB)
		if err != nil {
			return nil, dberr(err)
		}
		reader = d
		writers.Append(d)

	case "elastic":
		e, err := elastic.NewElastic(ctx, conf.Elastic)
		if err != nil {
			return nil, dberr(err)
		}
		reader = e
		nodes = e
		queue = e
		writers.Append(e)

	case "mongodb":
		m, err := mongodb.NewMongoDB(conf.MongoDB)
		if err != nil {
			return nil, dberr(err)
		}
		reader = m
		nodes = m
		queue = m
		writers.Append(m)

	default:
		return nil, fmt.Errorf("unknown database: '%s'", conf.Database)
	}

	fmt.Println(conf.Compute)
	switch strings.ToLower(conf.Compute) {
	case "manual":
		if nodes == nil {
			return nil, fmt.Errorf(
				"cannot enable manual compute backend, database %s does not implement "+
					"the scheduler service", conf.Database)
		}
		if queue == nil {
			return nil, fmt.Errorf(
				"cannot enable manual compute backend, database %s does not implement "+
					"a task queue", conf.Database)
		}

		sched = &scheduler.Scheduler{
			Conf:  conf.Scheduler,
			Log:   log.Sub("scheduler"),
			Nodes: nodes,
			Queue: queue,
			Event: &writers,
		}

	case "aws-batch":
		b, err := batch.NewBackend(conf.AWSBatch)
		if err != nil {
			return nil, err
		}
		writers.Append(b)

	case "gridengine":
		writers.Append(gridengine.NewBackend(conf))
	case "htcondor":
		writers.Append(htcondor.NewBackend(conf))
	case "local":
		writers.Append(local.NewBackend(conf, workerCmd.Run, log.Sub("local")))
	case "noop":
		writers.Append(noop.NewBackend())
	case "pbs":
		writers.Append(pbs.NewBackend(conf))
	case "slurm":
		writers.Append(slurm.NewBackend(conf))
	case "", "none":
		// Do nothing.
	default:
		return nil, fmt.Errorf("unknown compute backend: '%s'", conf.Compute)
	}

	var writer events.Writer
	var err error

	eventWriterSet := make(map[string]interface{})
	for _, w := range conf.EventWriters {
		eventWriterSet[strings.ToLower(w)] = nil
	}

	for e := range eventWriterSet {
		switch strings.ToLower(e) {
		case strings.ToLower(conf.Database):
			// noop
		case "log":
			// noop
		case "boltdb":
			writer, err = boltdb.NewBoltDB(conf.BoltDB)
		case "dynamodb":
			writer, err = dynamodb.NewDynamoDB(conf.DynamoDB)
		case "elastic":
			writer, err = elastic.NewElastic(ctx, conf.Elastic)
		case "kafka":
			writer, err = events.NewKafkaWriter(ctx, conf.Kafka)
		case "mongodb":
			writer, err = mongodb.NewMongoDB(conf.MongoDB)
		default:
			return nil, fmt.Errorf("unknown event writer: '%s'", e)
		}
		if err != nil {
			return nil, fmt.Errorf("error occurred while initializing the %s event writer: %v", e, err)
		}
		if writer != nil {
			writers.Append(writer)
		}
	}

	return &Server{
		Server: &server.Server{
			RPCAddress:       ":" + conf.Server.RPCPort,
			HTTPPort:         conf.Server.HTTPPort,
			Password:         conf.Server.Password,
			DisableHTTPCache: conf.Server.DisableHTTPCache,
			Log:              log,
			Tasks: &server.TaskService{
				Name:  conf.Server.ServiceName,
				Event: &writers,
				Read:  reader,
			},
			Events: &events.Service{Writer: &writers},
			Nodes:  nodes,
		},
		Scheduler: sched,
	}, nil
}

// Run runs a default Funnel server.
// This opens a database, and starts an API server, scheduler and task logger.
// This blocks indefinitely.
func (s *Server) Run(ctx context.Context) error {

	// Start server
	errch := make(chan error)
	go func() {
		errch <- s.Server.Serve(ctx)
	}()

	// Start Scheduler
	if s.Scheduler != nil {
		go func() {
			errch <- s.Scheduler.Run(ctx)
		}()
	}

	// Block until done.
	// Server and scheduler must be stopped via the context.
	return <-errch
}

func dberr(err error) error {
	return fmt.Errorf("error occurred while connecting to or creating the database: %v", err)
}
