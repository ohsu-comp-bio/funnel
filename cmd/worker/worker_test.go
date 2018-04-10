package worker

import (
	"context"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/ohsu-comp-bio/funnel/cmd/util"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
)

func TestPersistentPreRun(t *testing.T) {
	host := "test"
	rpcport := "9999"

	cwd, _ := os.Getwd()
	workDir := path.Join(cwd, "funnel-work-dir")

	fileConf := config.DefaultConfig()
	tmp, cleanup := util.TempConfigFile(fileConf, "testconfig.yaml")
	defer cleanup()

	c, h := newCommandHooks()
	h.Run = func(ctx context.Context, conf config.Config, log *logger.Logger, taskID string) error {
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

	c.SetArgs([]string{
		"run", "--config", tmp, "--Server.HostName", "test",
		"--Server.RPCPort", "9999", "--Server.RPCClientTimeout", "10ms",
		"--taskID", "test1234",
	})
	err := c.Execute()
	if err != nil {
		// since there is no server the rpc event writer will fail to connect
		// within its context deadline. We can ignore this error
		if !strings.Contains(err.Error(), "failed to instantiate EventWriter: context deadline exceeded") {
			t.Fatal(err)
		}
	}
}
