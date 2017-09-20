package util

import (
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
)

// MergeConfigFileWithFlags is a util used by server commands that use flags to set
// Funnel config values. These commands can also take in the path to a Funnel config file.
// This function ensures that the config gets set up properly. Flag values override values in
// the provided config file.
func MergeConfigFileWithFlags(file string, flagConf config.Config) (config.Config, error) {
	// parse config file if it exists
	conf := config.DefaultConfig()
	config.ParseFile(file, &conf)

	// make sure server address and password is inherited by scheduler nodes and workers
	conf = config.EnsureServerProperties(conf)
	flagConf = config.EnsureServerProperties(flagConf)

	// file vals <- cli val
	err := mergo.MergeWithOverwrite(&conf, flagConf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}
