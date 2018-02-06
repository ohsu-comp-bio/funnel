package tests

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"testing"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

var configFile string

func init() {
	// nanoseconds are important because the tests run faster than a millisecond
	// which can cause port conflicts
	rand.Seed(time.Now().UTC().UnixNano())
	flag.StringVar(&configFile, "funnel-config", configFile, "Funnel config file. Must be an absolute path.")
	flag.Parse()
}

// DefaultConfig returns a default configuration useful for testing,
// including temp storage dirs, random ports, S3 + minio config, etc.
func DefaultConfig() config.Config {
	conf := config.DefaultConfig()
	conf.AmazonS3.Disabled = true
	conf.GoogleStorage.Disabled = true
	conf.Swift.Disabled = true

	// Get config from test command line flag, if present.
	if configFile != "" {
		err := config.ParseFile(configFile, &conf)
		if err != nil {
			panic(err)
		}
	}

	return TestifyConfig(conf)
}

// TestifyConfig modifies a ports, directory paths, etc. to avoid
// conflicts between tests.
func TestifyConfig(conf config.Config) config.Config {
	conf = TempDirConfig(conf)
	conf = RandomPortConfig(conf)

	conf.Scheduler.ScheduleRate = time.Millisecond * 500
	conf.Node.UpdateRate = time.Millisecond * 1300
	conf.Worker.LogUpdateRate = time.Millisecond * 500
	conf.Worker.PollingRate = time.Millisecond * 100

	prefix := "funnel-e2e-tests-" + RandomString(6)
	conf.Elastic.IndexPrefix = prefix
	conf.MongoDB.Database = prefix
	conf.DynamoDB.TableBasename = prefix
	conf.Datastore.Project = prefix

	reconcile := time.Second * 5
	conf.HTCondor.ReconcileRate = reconcile
	conf.Slurm.ReconcileRate = reconcile
	conf.PBS.ReconcileRate = reconcile
	conf.GridEngine.ReconcileRate = reconcile
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

// SetLogOutput provides a method for connecting funnel logso the the test logger
func SetLogOutput(l *logger.Logger, t *testing.T) {
	l.SetOutput(TestingWriter(t))
}

// TestingWriter returns an io.Writer that writes each line via t.Log
func TestingWriter(t *testing.T) io.Writer {
	reader, writer := io.Pipe()
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			// Carriage return removes testing's file:line number and indent.
			// In this case, the file and line will always be "utils.go:62".
			// Go 1.9 introduced t.Helper() to fix this, but something about
			// this function being in a goroutine seems to break that.
			// Carriage return is the hack for now.
			t.Log("\r" + scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			t.Error("testing writer scanner error", err)
		}
	}()
	return writer
}

// LogConfig returns logger configuration useful for tests, which has a text indent.
func LogConfig() logger.Config {
	conf := logger.DefaultConfig()
	conf.TextFormat.ForceColors = true
	conf.TextFormat.Indent = "        "
	return conf
}
