package perf

import (
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"sync"
	"testing"
)

func BenchmarkRun5000NoWorkers(b *testing.B) {
	fun := e2e.NewFunnel()
	// No workers connected in this test
	fun.Conf.Scheduler = "manual"
	fun.Conf.LogLevel = "error"
	fun.StartServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j := 0; j < 5000; j++ {
			fun.Run(`
        --cmd 'echo'
      `)
		}
	}
}

func BenchmarkRun5000ConcurrentNoWorkers(b *testing.B) {
	fun := e2e.NewFunnel()
	// No workers connected in this test
	fun.Conf.Scheduler = "manual"
	fun.Conf.LogLevel = "error"
	fun.StartServer()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 5000; j++ {
			wg.Add(1)
			go func() {
				wg.Done()
				fun.Run(`
          --cmd 'echo'
        `)
			}()
		}
		wg.Wait()
	}
}
