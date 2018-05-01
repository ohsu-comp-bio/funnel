package tests

import (
	"flag"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/config/testconfig"
)

var configFile string

func init() {
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

	return testconfig.TestifyConfig(conf)
}
