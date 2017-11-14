package core

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"os"
	"testing"
)

var fun *e2e.Funnel
var log = logger.NewLogger("funnel-e2e-tes", logger.DefaultConfig())

func TestMain(m *testing.M) {
	fun = e2e.NewFunnel(e2e.DefaultConfig())
	fun.StartServer()
	e := m.Run()
	os.Exit(e)
}
