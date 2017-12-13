package node

import (
	"context"
	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/util"
	"syscall"
)

// Run runs a node with the given config, blocking until the node exits.
func Run(conf config.Config) error {
	log := logger.NewLogger("node", conf.Logger)

	ctx, cancel := context.WithCancel(context.Background())
	ctx = util.SignalContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if conf.Node.ID == "" {
		conf.Node.ID = scheduler.GenNodeID("manual")
	}

	ctx := context.Background()
	ctx = util.SignalContext(ctx, syscall.SIGINT, syscall.SIGTERM)

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
