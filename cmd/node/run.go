package node

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/compute/scheduler"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/util"
	"syscall"
)

// Run runs a node with the given config, blocking until the node exits.
func Run(conf config.Config) error {
	log := logger.NewLogger("node", conf.Scheduler.Node.Logger)

	if conf.Scheduler.Node.ID == "" {
		conf.Scheduler.Node.ID = scheduler.GenNodeID("manual")
	}

	n, err := scheduler.NewNode(conf, log)
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx = util.SignalContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	n.Run(ctx)

	return nil
}
