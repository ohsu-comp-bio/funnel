package worker

import (
	"context"
	"fmt"
	"strings"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/server/datastore"
	"github.com/ohsu-comp-bio/funnel/server/dynamodb"
	"github.com/ohsu-comp-bio/funnel/server/elastic"
	"github.com/ohsu-comp-bio/funnel/server/mongodb"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/worker"
)

// Run runs the "worker run" command.
func Run(ctx context.Context, conf config.Config, log *logger.Logger, taskID string) error {
	w, err := NewWorker(ctx, conf, log)
	if err != nil {
		return err
	}
	return w.Run(ctx, taskID)
}

// NewWorker returns a new Funnel worker based on the given config.
func NewWorker(ctx context.Context, conf config.Config, log *logger.Logger) (*worker.DefaultWorker, error) {
	log.Debug("NewWorker", "config", conf)

	var err error
	var db tes.ReadOnlyServer
	var reader worker.TaskReader
	var writer events.Writer
	var writers events.MultiWriter

	eventWriterSet := map[string]interface{}{
		strings.ToLower(conf.Database): nil,
	}
	for _, w := range conf.EventWriters {
		eventWriterSet[strings.ToLower(w)] = nil
	}

	for e := range eventWriterSet {
		switch e {
		case "log":
			writer = &events.Logger{Log: log}
		case "boltdb":
			writer, err = events.NewRPCWriter(ctx, conf.Server)
		case "dynamodb":
			writer, err = dynamodb.NewDynamoDB(conf.DynamoDB)
			writer = &events.Retrier{MaxRetrier: &util.MaxRetrier{MaxTries: 10}, Writer: writer}
		case "datastore":
			writer, err = datastore.NewDatastore(conf.Datastore)
		case "elastic":
			writer, err = elastic.NewElastic(conf.Elastic)
		case "kafka":
			writer, err = events.NewKafkaWriter(ctx, conf.Kafka)
		case "pubsub":
			writer, err = events.NewPubSubWriter(ctx, conf.PubSub)
		case "mongodb":
			writer, err = mongodb.NewMongoDB(conf.MongoDB)
		default:
			err = fmt.Errorf("unknown event writer: %s", e)
		}
		if err != nil {
			return nil, fmt.Errorf("error occurred while initializing the %s event writer: %v", e, err)
		}
		if writer != nil {
			writers = append(writers, writer)
		}
	}

	writer = &events.SystemLogFilter{Writer: &writers, Level: conf.Logger.Level}
	writer = &events.ErrLogger{Writer: writer, Log: log}

	switch strings.ToLower(conf.Database) {
	case "datastore":
		db, err = datastore.NewDatastore(conf.Datastore)
	case "dynamodb":
		db, err = dynamodb.NewDynamoDB(conf.DynamoDB)
	case "elastic":
		db, err = elastic.NewElastic(conf.Elastic)
	case "mongodb":
		db, err = mongodb.NewMongoDB(conf.MongoDB)
	case "boltdb":
		reader, err = worker.NewRPCTaskReader(ctx, conf.Server)
	default:
		err = fmt.Errorf("unknown database: '%s'", conf.Database)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate database client: %v", err)
	}
	if reader == nil {
		reader = worker.NewGenericTaskReader(db.GetTask)
	}

	store, err := storage.NewStorage(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate Storage backend: %v", err)
	}

	w := &worker.DefaultWorker{
		Conf:        conf.Worker,
		Store:       store,
		TaskReader:  reader,
		EventWriter: writer,
	}

	return w, nil
}
