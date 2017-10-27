package kafka

import (
	workerCmd "github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"testing"
)

func TestKafkaWorkerRun(t *testing.T) {

	c := e2e.DefaultConfig()
	c.Backend = "noop"
	c.Worker.ActiveEventWriters = []string{"kafka", "log", "rpc"}
	c.Worker.EventWriters.Kafka.Servers = []string{"localhost:9092"}

	log := logger.NewLogger("kafka-worker-test", c.Server.Logger)

	f := e2e.NewFunnel(c)
	f.StartServer()

	task := &tes.Task{}
	b := events.TaskBuilder{Task: task}
	l := &events.Logger{Log: log}
	m := events.MultiWriter(b, l)

	r, err := events.NewKafkaReader(c.Worker.EventWriters.Kafka, m)
	defer r.Close()
	if err != nil {
		t.Fatal(err)
	}

	// this only writes the task to the DB since the 'noop'
	// compute backend is in use
	id := f.Run(`
    --sh 'echo hello world'
  `)

	err = workerCmd.Run(c.Worker, id, log)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	f.Wait(id)

	if task.State != tes.State_COMPLETE {
		t.Fatal("unexpected state")
	}
}
