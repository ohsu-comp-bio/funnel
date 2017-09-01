package e2e

import (
	"os"
	"testing"
)

var fun *Funnel

func TestMain(m *testing.M) {
	// logging setup in utils.go
	fun = NewFunnel(DefaultConfig())
	fun.StartServer()
	e := m.Run()
	os.Exit(e)
}
