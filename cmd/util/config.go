package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/imdario/mergo"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/spf13/pflag"
)

func normalize(name string) string {
	from := []string{"-", "_"}
	to := "."
	for _, sep := range from {
		name = strings.Replace(name, sep, to, -1)
	}
	return strings.ToLower(name)
}

// NormalizeFlags allows for flags to be case and separator insensitive.
// Use it by passing it to cobra.Command.SetGlobalNormalizationFunc
func NormalizeFlags(f *pflag.FlagSet, name string) pflag.NormalizedName {
	lookup := map[string]string{"help": "help", normalize(name): name}

	f.VisitAll(func(f *pflag.Flag) {
		lookup[normalize(f.Name)] = f.Name
	})

	return pflag.NormalizedName(lookup[normalize(name)])
}

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

	// file vals <- cli val
	err = mergo.MergeWithOverwrite(&conf, flagConf)
	if err != nil {
		return conf, err
	}

	defaults := config.DefaultConfig()
	if conf.Server.RPCAddress() != defaults.Server.RPCAddress() {
		if conf.Server.RPCAddress() != conf.RPCClient.ServerAddress {
			conf.RPCClient.ServerAddress = conf.Server.RPCAddress()
		}
	}

	return conf, nil
}

// TempConfigFile writes the configuration to a temporary file.
// Returns:
// - "path" is the path of the file.
// - "cleanup" can be called to remove the temporary file.
func TempConfigFile(c config.Config, name string) (path string, cleanup func()) {
	tmpdir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}

	cleanup = func() {
		os.RemoveAll(tmpdir)
	}

	p := filepath.Join(tmpdir, name)
	err = config.ToYamlFile(c, p)
	if err != nil {
		panic(err)
	}
	return p, cleanup
}
