package util

import (
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
)

func TestMergeConfigFileWithFlags(t *testing.T) {
	defaultConf := config.DefaultConfig()
	flagConf := config.Config{
		Server: config.Server{
			HostName: "test",
			RPCPort:  "9999",
		},
	}
	serverAddress := flagConf.Server.RPCAddress()

	result, err := MergeConfigFileWithFlags("", flagConf)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if result.Server.RPCAddress() != serverAddress {
		t.Error("unexpected server address")
	}
	if result.Server.HTTPPort != defaultConf.Server.HTTPPort {
		t.Error("expected Config.Server.HTTPPort to equal the value from from config.DefaultValue()")
	}
	if result.Compute != defaultConf.Compute {
		t.Error("expected Config.Compute to equal default value from config.DefaultValue()")
	}

	fileConf := config.Config(defaultConf)
	fileConf.Server.HTTPPort = "8888"
	tmp, cleanup := config.ToYamlTempFile(fileConf, "testconfig.yaml")
	defer cleanup()
	result, err = MergeConfigFileWithFlags(tmp, flagConf)
	if err != nil {
		t.Error("unexpected error", err)
	}
	if result.Server.RPCAddress() != serverAddress {
		t.Error("unexpected server address")
	}
	if defaultConf.Server.HTTPPort == fileConf.Server.HTTPPort {
		t.Error("ERROR")
	}
	if result.Server.HTTPPort != fileConf.Server.HTTPPort {
		t.Error("expected Config.Server.HTTPPort to equal the value from the config file")
	}
	if result.Compute != defaultConf.Compute {
		t.Error("expected Config.Compute to equal default value from config.DefaultValue()")
	}
}
