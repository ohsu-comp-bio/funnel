package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/worker"
)

// Run configures and runs a Worker
func Run(conf config.Worker, taskID string, log *logger.Logger) error {
	w, err := worker.NewDefaultWorker(conf, taskID, log)
	if err != nil {
		return err
	}
	w.Run(context.Background())
	return nil
}
