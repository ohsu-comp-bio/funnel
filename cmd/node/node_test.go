package node

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"os"
	"path"
	"testing"
)

func TestPersistentPreRun(t *testing.T) {
	serverAddress := "test:9999"

	cwd, _ := os.Getwd()
	workDir := path.Join(cwd, "funnel-work-dir")

	fileConf := config.DefaultConfig()
	tmp, cleanup := fileConf.ToYamlTempFile("testconfig.yaml")
	defer cleanup()

	c, h := newCommandHooks()
	h.Run = func(conf config.Config) error {
		if conf.Scheduler.Node.ServerAddress != serverAddress {
			t.Fatal("unexpected ServerAddress in node config")
		}
		if conf.Scheduler.Node.WorkDir != workDir {
			t.Fatal("unexpected WorkDir in node config")
		}

		if conf.Worker.WorkDir != workDir {
			t.Fatal("unexpected WorkDir in node config")
		}
		if conf.Worker.EventWriters.RPC.ServerAddress != serverAddress {
			t.Fatal("unexpected ServerAddress in worker config")
		}
		if conf.Worker.TaskReaders.RPC.ServerAddress != serverAddress {
			t.Fatal("unexpected ServerAddress in worker config")
		}

		return nil
	}

	c.SetArgs([]string{"run", "--config", tmp, "--server-address", serverAddress})
	c.Execute()
}
