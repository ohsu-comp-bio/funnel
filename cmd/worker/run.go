package worker

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/server/dynamodb"
	"github.com/ohsu-comp-bio/funnel/server/elastic"
	"github.com/ohsu-comp-bio/funnel/server/mongodb"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/worker"
	"path"
)

// Run configures and runs a Worker
func Run(conf config.Worker, taskID string, log *logger.Logger) error {
	w, err := NewDefaultWorker(conf, taskID, log)
	if err != nil {
		return err
	}
	w.Run(context.Background())
	return nil
}

// NewDefaultWorker returns a new configured DefaultWorker instance.
func NewDefaultWorker(conf config.Worker, taskID string, log *logger.Logger) (worker.Worker, error) {

	var err error
	var db tes.TaskServiceServer
	var reader worker.TaskReader
	var writer events.Writer

	// Map files into this baseDir
	baseDir := path.Join(conf.WorkDir, taskID)

	err = util.EnsureDir(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create worker baseDir: %v", err)
	}

	switch conf.TaskReader {
	case "rpc":
		reader, err = worker.NewRPCTaskReader(conf.TaskReaders.RPC, taskID)
	case "dynamodb":
		db, err = dynamodb.NewDynamoDB(conf.TaskReaders.DynamoDB)
	case "elastic":
		db, err = elastic.NewTES(conf.EventWriters.Elastic)
	case "mongodb":
		db, err = mongodb.NewMongoDB(conf.TaskReaders.MongoDB)
	default:
		err = fmt.Errorf("unknown TaskReader")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate TaskReader: %v", err)
	}

	if reader == nil {
		reader = worker.NewGenericTaskReader(db.GetTask, taskID)
	}

	writers := []events.Writer{}
	for _, w := range conf.ActiveEventWriters {
		switch w {
		case "log":
			writer = &events.Logger{Log: log}
		case "rpc":
			writer, err = events.NewRPCWriter(conf.EventWriters.RPC)
		case "dynamodb":
			writer, err = dynamodb.NewDynamoDB(conf.EventWriters.DynamoDB)
		case "elastic":
			writer, err = elastic.NewElastic(conf.EventWriters.Elastic)
		case "mongodb":
			writer, err = mongodb.NewMongoDB(conf.EventWriters.MongoDB)
		case "kafka":
			writer, err = events.NewKafkaWriter(conf.EventWriters.Kafka)
		default:
			err = fmt.Errorf("unknown EventWriter")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate EventWriter: %v", err)
		}
		writers = append(writers, writer)
	}

	m := events.MultiWriter(writers...)
	ew := &events.ErrLogger{Writer: m, Log: log}

	return &worker.DefaultWorker{
		Conf:       conf,
		Mapper:     worker.NewFileMapper(baseDir),
		Store:      storage.Storage{},
		TaskReader: reader,
		Event:      events.NewTaskWriter(taskID, 0, conf.Logger.Level, ew),
	}, nil
}
