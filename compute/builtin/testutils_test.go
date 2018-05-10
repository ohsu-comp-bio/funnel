package builtin

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/config/testconfig"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/util/rpc"
	"google.golang.org/grpc"
)

func testConfig() config.Config {
	conf := config.DefaultConfig()
	conf = testconfig.TestifyConfig(conf)
	conf.Compute = "builtin"
	conf.Node.UpdateRate = config.Duration(10 * time.Millisecond)
	conf.Scheduler.NodePingTimeout = config.Duration(2 * time.Second)
	conf.Scheduler.NodeDeadTimeout = config.Duration(2 * time.Second)

	workDir, err := ioutil.TempDir("", "funnel-test-node-")
	if err != nil {
		panic(err)
	}

	conf.Worker.WorkDir = workDir
	conf.Node.WorkDir = workDir
	return conf
}

type testNode struct {
	*NodeProcess
	conn *grpc.ClientConn
	done chan struct{}
}

func (t *testNode) Start() context.CancelFunc {
	t.done = make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		err := t.NodeProcess.Run(ctx)
		if err != nil {
			panic(err)
		}
		close(t.done)
	}()
	time.Sleep(500 * time.Millisecond)
	return cancel
}

func (t *testNode) Wait() {
	<-t.done
}

func newTestNode(conf config.Config) *testNode {
	log := logger.NewLogger("test-node", logger.DebugConfig())

	// Create client for scheduler RPC
	ctx := context.Background()
	conn, err := rpc.Dial(ctx, conf.Server)
	if err != nil {
		panic(fmt.Errorf("connecting to server: %s", err))
	}
	client := NewSchedulerServiceClient(conn)

	node, err := NewNodeProcess(conf.Node, client, NoopWorker, log)
	if err != nil {
		panic(err)
	}
	return &testNode{NodeProcess: node, conn: conn}
}

type testSched struct {
	*Scheduler
	srv *grpc.Server
}

func newTestSched(conf config.Config) *testSched {
	log := logger.NewLogger("test-sched", logger.DebugConfig())

	// Open TCP connection for RPC
	lis, err := net.Listen("tcp", conf.Server.RPCAddress())
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	ev := &events.Logger{Log: log}
	sched, err := NewScheduler(conf.Scheduler, log, ev)
	if err != nil {
		panic(err)
	}

	RegisterSchedulerServiceServer(grpcServer, sched)
	if err != nil {
		panic(err)
	}

	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	// Give the scheduler server time to start.
	time.Sleep(150 * time.Millisecond)

	return &testSched{Scheduler: sched, srv: grpcServer}
}

func timeLimit(name string, d time.Duration) func() {
	stop := make(chan struct{})

	go func() {
		select {
		case <-time.After(d):
			panic(fmt.Sprintf("time limit expired for %s", name))
		case <-stop:
		}
	}()

	return func() {
		close(stop)
	}
}
