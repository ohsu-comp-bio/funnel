package worker

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
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
	h.Run = func(ctx context.Context, conf config.Config, taskID string, log *logger.Logger) error {
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

	c.SetArgs([]string{"run", "--config", tmp, "--Server.HostName", "test", "--Server.RPCPort", "9999", "--taskID", "test1234"})
	err := c.Execute()
	if err != nil {
		t.Fatal(err)
	}
}
