package util

import (
	"fmt"
	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"strings"
)

// MergeConfigFileWithFlags is a util used by server commands that use flags to set
// Funnel config values. These commands can also take in the path to a Funnel config file.
// This function ensures that the config gets set up properly. Flag values override values in
// the provided config file.
func MergeConfigFileWithFlags(file string, flagConf config.Config) (config.Config, error) {
	// parse config file if it exists
	conf := config.DefaultConfig()
	err := config.ParseFile(file, &conf)
	if err != nil {
		return conf, err
	}

	// make sure server address and password is inherited by scheduler nodes and workers
	conf = config.EnsureServerProperties(conf)
	flagConf = config.EnsureServerProperties(flagConf)

	// file vals <- cli val
	err = mergo.MergeWithOverwrite(&conf, flagConf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

// ParseServerAddressFlag parses a gRPC server address and sets the relevant
// fields in the config
func ParseServerAddressFlag(address string, conf config.Config) (config.Config, error) {
	if address != "" {
		parts := strings.Split(address, ":")
		if len(parts) != 2 {
			return conf, fmt.Errorf("error parsing server address")
		}
		conf.Server.HostName = parts[0]
		conf.Server.RPCPort = parts[1]
	}
	return conf, nil
}
