package util

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"testing"
)

func TestMergeConfigFileWithFlags(t *testing.T) {
	serverAddress := "test:9090"

	flagConf := config.Config{}
	flagConf.Scheduler.Node.ServerAddress = serverAddress
	result, err := MergeConfigFileWithFlags("", flagConf)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Scheduler.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.Worker.EventWriters.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected  server address in worker config")
	}
	if result.Worker.TaskReaders.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected  server address in worker config")
	}

	fileConf := config.DefaultConfig()
	tmp, cleanup := fileConf.ToYamlTempFile("testconfig.yaml")
	defer cleanup()
	result, err = MergeConfigFileWithFlags(tmp, flagConf)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Scheduler.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.Worker.EventWriters.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected  server address in worker config")
	}
	if result.Worker.TaskReaders.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected  server address in worker config")
	}

	if result.Backend == "" {
		t.Fatal("expected Config.Backend to equal default value from config.DefaultValue()")
	}
}
