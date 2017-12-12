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
	tmp, cleanup := config.ToYamlTempFile(fileConf, "testconfig.yaml")
	defer cleanup()

	c, h := newCommandHooks()
	h.Run = func(conf config.Config) error {
		if conf.Node.ServerAddress != serverAddress {
			t.Fatal("unexpected ServerAddress in Node config", conf.Node.ServerAddress)
		}
		if conf.Worker.WorkDir != workDir {
			t.Fatal("unexpected WorkDir in Worker config", conf.Worker.WorkDir)
		}
		if conf.RPC.ServerAddress != serverAddress {
			t.Fatal("unexpected ServerAddress in RPC config", conf.RPC.ServerAddress)
		}

		return nil
	}

	c.SetArgs([]string{"run", "--config", tmp, "--server", serverAddress})
	c.Execute()
}
