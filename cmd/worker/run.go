package worker

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
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
		reader, err = worker.NewRPCTaskReader(conf, taskID)
	case "dynamodb":
		reader, err = worker.NewDynamoDBTaskReader(conf.TaskReaders.DynamoDB, taskID)
		// case "mongodb":
		//   reader, err = worker.NewMongoDBTaskReader(conf.TaskReaders.MongoDB, taskID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate TaskReader: %v", err)
	}

	writers := []events.Writer{}
	for _, w := range conf.ActiveEventWriters {
		switch w {
		case "dynamodb":
			writer, err = events.NewDynamoDBEventWriter(conf.EventWriters.DynamoDB)
		case "log":
			writer = &events.Logger{Log: log}
		case "rpc":
			writer, err = events.NewRPCWriter(conf)
		case "elastic":
			writer, err = elastic.NewElastic(conf.EventWriters.Elastic)
		case "mongodb":
			writer, err = mongodb.NewMongoDB(conf.EventWriters.MongoDB)
		default:
			err = fmt.Errorf("unknown EventWriter")
		}
		if err != nil {
			return nil, fmt.Errorf("failed to instantiate EventWriter: %v", err)
		}
		writers = append(writers, writer)
	}

	return &worker.DefaultWorker{
		Conf:       conf,
		Mapper:     worker.NewFileMapper(baseDir),
		Store:      storage.Storage{},
		TaskReader: reader,
		Event:      events.NewTaskWriter(taskID, 0, conf.Logger.Level, events.MultiWriter(writers...)),
	}, nil
}
