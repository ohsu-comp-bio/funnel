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
	serverAddress := "test:9999"

	cwd, _ := os.Getwd()
	workDir := path.Join(cwd, "funnel-work-dir")

	fileConf := config.DefaultConfig()
	tmp, cleanup := fileConf.ToYamlTempFile("testconfig.yaml")
	defer cleanup()

	c, h := newCommandHooks()
	h.Run = func(ctx context.Context, conf config.Worker, taskID string, l *logger.Logger) error {
		if conf.WorkDir != workDir {
			t.Fatal("unexpected WorkDir in worker config")
		}
		if conf.EventWriters.RPC.ServerAddress != serverAddress {
			t.Fatal("unexpected ServerAddress in worker config")
		}
		if conf.TaskReaders.RPC.ServerAddress != serverAddress {
			t.Fatal("unexpected ServerAddress in worker config")
		}
		return nil
	}

	c.SetArgs([]string{"run", "--config", tmp, "--server-address", serverAddress, "--task-id", "test1234"})
	c.Execute()
}
