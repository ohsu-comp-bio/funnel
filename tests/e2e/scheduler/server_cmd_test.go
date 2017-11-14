package scheduler

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/cmd/server"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"testing"
)

// Test a bug fix where running the server with the "manual" backend
// was causing a panic.
func TestServerRunManualPanic(t *testing.T) {
	conf := e2e.DefaultConfig()
	conf.Backend = "manual"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		server.Run(ctx, conf)
	}()

	conn, err := e2e.NewRPCConn(conf)
	if err != nil {
		t.Fatal(err)
	}

	cli := tes.NewTaskServiceClient(conn)

	// The bug was that the server had not properly configured the compute
	// backend, so creating a task would result in a nil pointer panic.
	cli.CreateTask(ctx, e2e.HelloWorld)
}
