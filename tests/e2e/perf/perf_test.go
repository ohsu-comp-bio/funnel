package perf

import (
	"context"
	tl "github.com/ohsu-comp-bio/funnel/proto/tasklogger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"google.golang.org/grpc"
	"sync"
	"testing"
	"time"
)

func BenchmarkRunSerialNoNodes(b *testing.B) {
	fun := e2e.NewFunnel(e2e.DefaultConfig())
	defer fun.Cleanup()
	// No nodes connected in this test
	fun.Conf.Backend = "manual"
	fun.Conf.Server.Logger.Level = "error"

	fun.StartServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fun.Run(`
      --cmd 'echo'
    `)
	}
}

func BenchmarkRunConcurrentNoNodes(b *testing.B) {
	fun := e2e.NewFunnel(e2e.DefaultConfig())
	defer fun.Cleanup()
	// No nodes connected in this test
	fun.Conf.Backend = "manual"
	fun.Conf.Server.Logger.Level = "error"
	fun.StartServer()
	b.ResetTimer()

	var wg sync.WaitGroup

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fun.Run(`
        --cmd 'echo'
      `)
		}()
	}
	wg.Wait()
}

func BenchmarkRunConcurrentWithFakeNodes(b *testing.B) {
	fun := e2e.NewFunnel(e2e.DefaultConfig())
	defer fun.Cleanup()
	// Nodes are simulated by goroutines writing to the scheduler API
	fun.Conf.Backend = "manual"
	fun.Conf.Server.Logger.Level = "error"
	fun.StartServer()

	var wg sync.WaitGroup
	ids := make(chan string, 1000)
	done := make(chan struct{})
	defer close(done)

	// Generate a 1000 character string to write as stdout logs
	content := ""
	for j := 0; j < 1000; j++ {
		content += "a"
	}

	// When a task is created, start a fake node that writes to the database.
	go func() {
		for {
			select {
			case id := <-ids:
				// fake node that writes to UpdateExecutorLogs every tick
				go func(id string) {
					conn, err := grpc.Dial(fun.Conf.Server.RPCAddress(), grpc.WithInsecure())
					if err != nil {
						panic(err)
					}
					cli := tl.NewTaskLoggerServiceClient(conn)
					_ = cli
					ticker := time.NewTicker(time.Millisecond * 20)

					for {
						select {
						case <-ticker.C:
							cli.UpdateExecutorLogs(context.Background(), &tl.UpdateExecutorLogsRequest{
								Id:   id,
								Step: 0,
								Log: &tes.ExecutorLog{
									Stdout: content,
								},
							})
						case <-done:
							return
						}
					}
				}(id)
			case <-done:
				return
			}
		}
	}()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ids <- fun.Run(`
        --cmd 'echo'
      `)
		}()
	}

	wg.Wait()
}
