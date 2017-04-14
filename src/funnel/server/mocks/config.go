package servermocks

import (
	"fmt"
	"funnel/config"
	"math/rand"
	"time"
)

// Config returns a new Config instance configured with
// values for testing (random port, temp database file, etc)
func Config(conf config.Config) config.Config {
	port := randomPort()
	conf.RPCPort = port
	return conf
}

func init() {
	// nanoseconds are important because the tests run faster than a millisecond
	// which can cause port conflicts
	rand.Seed(time.Now().UTC().UnixNano())
}
func randomPort() string {
	min := 10000
	max := 20000
	n := rand.Intn(max-min) + min
	return fmt.Sprintf("%d", n)
}
