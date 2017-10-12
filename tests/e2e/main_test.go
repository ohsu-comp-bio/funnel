package e2e

import (
	"flag"
	"github.com/ohsu-comp-bio/funnel/config"
	"os"
	"testing"
)

var fun *Funnel
var configFile string

func init() {
	flag.StringVar(&configFile, "funnel-config", configFile, "Funnel config file. Must be an absolute path.")
}

func TestMain(m *testing.M) {
	flag.Parse()

	conf := config.DefaultConfig()
	if configFile != "" {
		err := config.ParseFile(configFile, &conf)
		if err != nil {
			panic(err)
		}
	}
	conf = TestifyConfig(conf)

	fun = NewFunnel(conf)
	fun.StartServer()
	e := m.Run()
	os.Exit(e)
}
