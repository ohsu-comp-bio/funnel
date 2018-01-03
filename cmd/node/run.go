package node

import (
	"context"
	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
)

// Run runs a node with the given config, blocking until the node exits.
func Run(ctx context.Context, conf config.Config, log *logger.Logger) error {
	if conf.Node.ID == "" {
		conf.Node.ID = scheduler.GenNodeID("manual")
	}

	ew, err := workerCmd.NewWorkerEventWriter(ctx, conf, log)
	if err != nil {
		return err
	}

	workerFactory := func(ctx context.Context, taskID string) error {
		return workerCmd.Run(ctx, conf, taskID, ew, log)
	}

	n, err := scheduler.NewNode(ctx, conf, workerFactory, log)
	if err != nil {
		return err
	}

	n.Run(ctx)

	return nil
}
