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
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
	"github.com/ohsu-comp-bio/funnel/worker"
	"path"
)

// Run configures and runs a Worker
func Run(ctx context.Context, conf config.Worker, taskID string, log *logger.Logger) error {
	log.Debug("Run Worker", "config", conf, "taskID", taskID)

	var err error
	var db tes.ReadOnlyServer
	var reader worker.TaskReader
	var writer events.Writer

	// Map files into this baseDir
	baseDir := path.Join(conf.WorkDir, taskID)

	err = fsutil.EnsureDir(baseDir)
	if err != nil {
		return fmt.Errorf("failed to create worker baseDir: %v", err)
	}

	switch conf.TaskReader {
	case "rpc":
		reader, err = worker.NewRPCTaskReader(conf.TaskReaders.RPC, taskID)
	case "dynamodb":
		db, err = dynamodb.NewDynamoDB(conf.TaskReaders.DynamoDB)
	case "elastic":
		db, err = elastic.NewElastic(ctx, conf.EventWriters.Elastic)
	case "mongodb":
		db, err = mongodb.NewMongoDB(conf.TaskReaders.MongoDB)
	default:
		err = fmt.Errorf("unknown TaskReader")
	}
	if err != nil {
		return fmt.Errorf("failed to instantiate TaskReader: %v", err)
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
			writer, err = elastic.NewElastic(ctx, conf.EventWriters.Elastic)
		case "mongodb":
			writer, err = mongodb.NewMongoDB(conf.EventWriters.MongoDB)
		case "kafka":
			k, kerr := events.NewKafkaWriter(conf.EventWriters.Kafka)
			defer k.Close()
			err = kerr
			writer = k
		default:
			err = fmt.Errorf("unknown EventWriter")
		}
		if err != nil {
			return fmt.Errorf("failed to instantiate EventWriter: %v", err)
		}
		writers = append(writers, writer)
	}

	m := events.MultiWriter(writers)
	ew := &events.ErrLogger{Writer: &m, Log: log}

	w := &worker.DefaultWorker{
		Conf:       conf,
		Mapper:     worker.NewFileMapper(baseDir),
		Store:      storage.Storage{},
		TaskReader: reader,
		Event:      events.NewTaskWriter(taskID, 0, conf.Logger.Level, ew),
	}
	w.Run(ctx)
	return nil
}
