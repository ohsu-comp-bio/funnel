package testconfig

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

func init() {
	// nanoseconds are important because the tests run faster than a millisecond
	// which can cause port conflicts
	rand.Seed(time.Now().UTC().UnixNano())
}

// TestifyConfig modifies a ports, directory paths, etc. to avoid
// conflicts between tests.
func TestifyConfig(conf config.Config) config.Config {
	conf = TempDirConfig(conf)
	conf = RandomPortConfig(conf)

	conf.Node.UpdateRate = config.Duration(time.Millisecond * 1300)
	conf.Worker.LogUpdateRate = config.Duration(time.Millisecond * 500)
	conf.Worker.PollingRate = config.Duration(time.Millisecond * 100)

	prefix := "funnel-e2e-tests-" + RandomString(6)
	conf.Elastic.IndexPrefix = prefix
	conf.MongoDB.Database = prefix
	conf.DynamoDB.TableBasename = prefix
	conf.Datastore.Project = prefix

	reconcile := config.Duration(time.Second * 5)
	conf.HTCondor.ReconcileRate = reconcile
	conf.Slurm.ReconcileRate = reconcile
	conf.PBS.ReconcileRate = reconcile
	conf.AWSBatch.ReconcileRate = reconcile

	storageDir, _ := ioutil.TempDir("./test_tmp", "funnel-test-storage-")
	wd, _ := os.Getwd()
	fsutil.EnsureDir(storageDir)

	conf.LocalStorage = config.LocalStorage{
		AllowedDirs: []string{storageDir, wd},
	}

	return conf
}

// RandomPort returns a random port string between 10000 and 20000.
func RandomPort() string {
	min := 10000
	max := 40000
	n := rand.Intn(max-min) + min
	return fmt.Sprintf("%d", n)
}

// RandomPortConfig returns a modified config with random HTTP and RPC ports.
func RandomPortConfig(conf config.Config) config.Config {
	conf.Server.RPCPort = RandomPort()
	conf.Server.HTTPPort = RandomPort()
	return conf
}

// TempDirConfig returns a modified config with workdir and db path set to a temp. directory.
func TempDirConfig(conf config.Config) config.Config {
	os.Mkdir("./test_tmp", os.ModePerm)
	f, _ := ioutil.TempDir("./test_tmp", "funnel-test-")
	conf.Worker.WorkDir = f
	conf.BoltDB.Path = path.Join(f, "funnel.db")
	conf.Scheduler.DBPath = path.Join(f, "scheduler.db")
	return conf
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// LogConfig returns logger configuration useful for tests, which has a text indent.
func LogConfig() logger.Config {
	conf := logger.DefaultConfig()
	conf.TextFormat.ForceColors = true
	conf.TextFormat.Indent = "        "
	return conf
}
