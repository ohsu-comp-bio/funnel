package perf

import (
	"context"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"google.golang.org/grpc"
	"sync"
	"testing"
	"time"
)

func BenchmarkRunSerialNoWorkers(b *testing.B) {
	fun := e2e.NewFunnel()
	defer fun.Cleanup()
	// No workers connected in this test
	fun.Conf.Scheduler = "manual"
	fun.Conf.Logger.Level = "error"
	fun.StartServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fun.Run(`
      --cmd 'echo'
    `)
	}
}

func BenchmarkRunConcurrentNoWorkers(b *testing.B) {
	fun := e2e.NewFunnel()
	defer fun.Cleanup()
	// No workers connected in this test
	fun.Conf.Scheduler = "manual"
	fun.Conf.Logger.Level = "error"
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

func BenchmarkRunConcurrentWithFakeWorkers(b *testing.B) {
	fun := e2e.NewFunnel()
	defer fun.Cleanup()
	// Workers are simulated by goroutines writing to the scheduler API
	fun.Conf.Scheduler = "manual"
	fun.Conf.Logger.Level = "error"
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

	// When a task is created, start a fake worker that writes to the database.
	go func() {
		for {
			select {
			case id := <-ids:
				// fake worker that writes to UpdateExecutorLogs every tick
				go func(id string) {
					conn, err := grpc.Dial(fun.Conf.RPCAddress(), grpc.WithInsecure())
					if err != nil {
						panic(err)
					}
					s := pbf.NewSchedulerServiceClient(conn)
					_ = s
					ticker := time.NewTicker(time.Millisecond * 20)

					for {
						select {
						case <-ticker.C:
							s.UpdateExecutorLogs(context.Background(), &pbf.UpdateExecutorLogsRequest{
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
