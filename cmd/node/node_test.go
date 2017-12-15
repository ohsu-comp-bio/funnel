package node

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"os"
	"path"
	"testing"
)

func TestPersistentPreRun(t *testing.T) {
	host := "test"
	rpcport := "9999"

	cwd, _ := os.Getwd()
	workDir := path.Join(cwd, "funnel-work-dir")

	fileConf := config.DefaultConfig()
	tmp, cleanup := config.ToYamlTempFile(fileConf, "testconfig.yaml")
	defer cleanup()

	c, h := newCommandHooks()
	h.Run = func(conf config.Config) error {
		if conf.Server.HostName != host {
			t.Fatal("unexpected Server.HostName in config", conf.Server.HostName)
		}
		if conf.Server.RPCPort != rpcport {
			t.Fatal("unexpected Server.RPCAddress in config", conf.Server.RPCPort)
		}
		if conf.Worker.WorkDir != workDir {
			t.Fatal("unexpected Worker.WorkDir in config", conf.Worker.WorkDir)
		}

		return nil
	}

	c.SetArgs([]string{"run", "--config", tmp, "--Server.HostName", host, "--Server.RPCPort", rpcport})
	err := c.Execute()
	if err != nil {
		t.Fatal(err)
	}
}
