package server

import (
	"os"
	"testing"

	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tests"
)

var fun *tests.Funnel
var log = logger.NewLogger("funnel-e2e-server", logger.DefaultConfig())

func TestMain(m *testing.M) {
	tests.ParseConfig()
	fun = tests.NewFunnel(tests.DefaultConfig())
	fun.StartServer()
	e := m.Run()
	os.Exit(e)
}
