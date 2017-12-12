package util

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"testing"
)

func TestMergeConfigFileWithFlags(t *testing.T) {
	serverAddress := "test:9999"

	flagConf := config.Config{}
	flagConf, err := ParseServerAddressFlag(serverAddress, flagConf)
	if err != nil {
		t.Fatal("unexpected error", err)
	}
	result, err := MergeConfigFileWithFlags("", flagConf)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in RPC config")
	}

	fileConf := config.DefaultConfig()
	tmp, cleanup := config.ToYamlTempFile(fileConf, "testconfig.yaml")
	defer cleanup()
	result, err = MergeConfigFileWithFlags(tmp, flagConf)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	if result.Server.RPCAddress() != serverAddress {
		t.Fatal("unexpected server address")
	}
	if result.Node.ServerAddress != serverAddress {
		t.Fatal("unexpected node server address")
	}
	if result.RPC.ServerAddress != serverAddress {
		t.Fatal("unexpected server address in worker config")
	}

	if result.Compute != "local" {
		t.Fatal("expected Config.Compute to equal default value from config.DefaultValue()")
	}
}
