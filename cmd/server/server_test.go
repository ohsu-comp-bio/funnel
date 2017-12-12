package server

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"testing"
)

func TestPersistentPreRun(t *testing.T) {
	hostname := "test"
	rpc := "9999"
	serverAddress := hostname + ":" + rpc
	backend := "test-backend"

	fileConf := config.DefaultConfig()
	tmp, cleanup := config.ToYamlTempFile(fileConf, "testconfig.yaml")
	defer cleanup()

	c, h := newCommandHooks()
	h.Run = func(ctx context.Context, conf config.Config) error {
		if conf.Server.RPCAddress() != serverAddress {
			t.Fatal("unexpected hostname or rpc port in server config")
		}
		if conf.Server.HTTPPort != fileConf.Server.HTTPPort {
			t.Fatal("unexpected http port in server config")
		}
		if conf.Node.ServerAddress != serverAddress {
			t.Fatal("unexpected ServerAddress in node config")
		}
		if conf.RPC.ServerAddress != serverAddress {
			t.Fatal("unexpected ServerAddress in worker config")
		}
		if conf.Compute != backend {
			t.Fatal("unexpected Backend in config")
		}
		return nil
	}

	c.SetArgs([]string{"run", "--config", tmp, "--Server.Hostname", hostname, "--Server.RPCPort", rpc, "--Compute", backend})
	c.Execute()
}
