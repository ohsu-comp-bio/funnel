package testutils

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"time"
)

func init() {
	// nanoseconds are important because the tests run faster than a millisecond
	// which can cause port conflicts
	rand.Seed(time.Now().UTC().UnixNano())
}

// RandomPort returns a random port string between 10000 and 20000.
func RandomPort() string {
	min := 10000
	max := 20000
	n := rand.Intn(max-min) + min
	return fmt.Sprintf("%d", n)
}

// RandomPortConfig returns a modified config with random HTTP and RPC ports.
func RandomPortConfig(conf config.Config) config.Config {
	conf.RPCPort = RandomPort()
	conf.HTTPPort = RandomPort()
	return conf
}

// TempDirConfig returns a modified config with workdir and db path set to a temp. directory.
func TempDirConfig(conf config.Config) config.Config {
	os.Mkdir("./test_tmp", os.ModePerm)
	f, _ := ioutil.TempDir("./test_tmp", "funnel-test-")
	conf.WorkDir = f
	conf.DBPath = path.Join(f, "funnel.db")
	return conf
}
