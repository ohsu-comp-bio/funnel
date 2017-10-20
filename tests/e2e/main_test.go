package e2e

import (
	"os"
	"testing"
)

var fun *Funnel

func TestMain(m *testing.M) {
	fun = NewFunnel(DefaultConfig())
	fun.StartServer()
	e := m.Run()
	os.Exit(e)
}
