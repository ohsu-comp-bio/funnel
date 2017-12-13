package server

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"testing"
)

func TestPersistentPreRun(t *testing.T) {
	host := "test"
	rpcport := "9999"
	serverAddress := host + ":" + rpcport
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
		if conf.Compute != backend {
			t.Fatal("unexpected Backend in config")
		}
		return nil
	}

	c.SetArgs([]string{"run", "--config", tmp, "--Server.HostName", host, "--Server.RPCPort", rpcport, "--Compute", backend})
	err := c.Execute()
	if err != nil {
		t.Fatal(err)
	}
}
