package aws

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/worker"
	"os"
)

// RunTask handles TaskRunner startup in the context
// of an AWS Batch environment. No Worker is created,
// the task goes directly to a TaskRunner.
//
// TES Task and Executor logs are written to stderr
// via a Funnel logger, which is written to AWS CloudWatch.
//
// Other configuration is pulled from the environment where possible,
// such as the task ID, storage auth, etc.
func runTask(task *tes.Task, conf config.Config) error {
	if conf.Worker.ID == "" {
		conf.Worker.ID = scheduler.GenWorkerID("funnel")
	}

	log.Debug("RUN TASK", task)

	task.Id = os.Getenv("AWS_BATCH_JOB_ID")

	// Task logs will be written to this logger and formatted to JSON.
	// This get written to AWS CloudWatchLogs, where the Funnel + AWS Batch
	// proxy reads the log data.
	log := logger.New("aws-runtask", "task", task.Id)
	log.Configure(logger.Config{
		Level:     "info",
		Formatter: "json",
		JSONFormat: logger.JSONFormatConfig{
			DisableTimestamp: true,
		},
	})

	runner := worker.DefaultRunner{
		Conf:   conf.Worker,
		Mapper: worker.NewFileMapper(conf.WorkDir),
		Store:  storage.Storage{},
		Svc: &taskService{
			worker.NewThinTaskLogger(log),
			log,
			task,
		},
		Log: log,
	}

	runner.Run(context.Background())

	return nil
}

// AWS Batch specific worker.TaskService implementation.
// Batch doesn't need to read task state, and it writes
// task/executor logs to stderr.
type taskService struct {
	worker.TaskLogger
	log  logger.Logger
	task *tes.Task
}

func (ts *taskService) Task() (*tes.Task, error) {
	return ts.task, nil
}

// AWS Batch controls the task state externally, so the runner
// doesn't need any info from State. Alway return RUNNING.
func (ts *taskService) State() tes.State {
	return tes.State_RUNNING
}

// AWS Batch controls the task state externally,
// so this method is a noop with some logging.
func (ts *taskService) SetState(s tes.State) error {
	ts.log.Info("SetState", "state", s)
	return nil
}
