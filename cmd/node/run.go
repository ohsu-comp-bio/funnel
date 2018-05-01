package node

import (
	"context"
	"fmt"

	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/compute/builtin"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/util/rpc"
)

// Run runs a node with the given config, blocking until the node exits.
func Run(ctx context.Context, conf config.Config, log *logger.Logger) error {

	w, err := workerCmd.NewWorker(ctx, conf, log)
	if err != nil {
		return fmt.Errorf("creating worker: %s", err)
	}

	conn, err := rpc.Dial(ctx, conf.Server)
	if err != nil {
		return fmt.Errorf("connecting to server: %s", err)
	}
	defer conn.Close()
	client := builtin.NewSchedulerServiceClient(conn)

	conf.Node.WorkDir = conf.Worker.WorkDir
	n, err := builtin.NewNodeProcess(conf.Node, client, w.Run, log)
	if err != nil {
		return fmt.Errorf("creating node: %s", err)
	}

	return n.Run(ctx)
}
