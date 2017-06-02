package htcondor

import (
	"flag"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"os"
	"testing"
)

var log = logger.New("e2e-htcondor")
var fun *e2e.Funnel
var runTest = flag.Bool("run-test", false, "run e2e tests with dockerized scheduler")

func TestMain(m *testing.M) {
	log.Configure(logger.DebugConfig())

	flag.Parse()
	if !*runTest {
		log.Info("Skipping htcondor e2e tests...")
		os.Exit(0)
	}

	fun = e2e.NewFunnel()
	fun.StartServerInDocker("ohsucompbio/htcondor:latest", "htcondor", []string{})
	defer fun.CleanupTestServerContainer()

	m.Run()
	return
}

func TestHelloWorld(t *testing.T) {
	id := fun.Run(`
    --cmd 'echo hello world'
  `)
	task := fun.Wait(id)

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("Missing stdout")
	}
}

func TestResourceRequest(t *testing.T) {
	id := fun.Run(`
    --cmd 'echo I need resources!' --cpu 1 --ram 2 --disk 5
  `)
	task := fun.Wait(id)

	if task.Logs[0].Logs[0].Stdout != "I need resources!\n" {
		t.Fatal("Missing stdout")
	}
}
