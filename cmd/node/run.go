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
	conf.Node.ID = scheduler.GenNodeID("node")

	w, err := workerCmd.NewWorker(ctx, conf, log)
	if err != nil {
		return err
	}

	n, err := scheduler.NewNodeProcess(conf, w.Run, log)
	if err != nil {
		return err
	}

	return n.Run(ctx)
}
