package core

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tests"
	"os"
	"testing"
)

var fun *tests.Funnel
var log = logger.NewLogger("funnel-e2e-core", logger.DefaultConfig())

func TestMain(m *testing.M) {
	fun = tests.NewFunnel(tests.DefaultConfig())
	fun.StartServer()
	e := m.Run()
	os.Exit(e)
}
