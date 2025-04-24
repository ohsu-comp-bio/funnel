package util

import (
	"os"
	"path/filepath"
	"reflect"

	"github.com/ohsu-comp-bio/funnel/config"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

/*
The proto merge function merges but does not replace durationpb values,
but the original intent of MergeConfigFile function is in places where both
configs are populated the new config replaces the value of the default config
so this function recursively traverses the config structures replacing the durationpb values.
*/
func mergeDurations(base, override proto.Message) {
	bv := reflect.ValueOf(base).Elem()
	ov := reflect.ValueOf(override).Elem()

	for i := range bv.NumField() {
		bf := bv.Field(i)
		of := ov.Field(i)

		// If the field is a pointer and set in override
		if of.Kind() == reflect.Ptr && !of.IsNil() {
			switch bf.Type().String() {
			case "*durationpb.Duration":
				overrideDur := of.Interface().(*durationpb.Duration)
				if overrideDur != nil && overrideDur.AsDuration() != 0 {
					bf.Set(reflect.ValueOf(overrideDur))
				}
			default:
				// Recurse into nested messages
				if of.Type().Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) &&
					bf.Type().Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) &&
					!bf.IsNil() {
					mergeDurations(bf.Interface().(proto.Message), of.Interface().(proto.Message))
				}
			}
		}
	}
}

// MergeConfigFileWithFlags is a util used by server commands that use flags to set
// Funnel config values. These commands can also take in the path to a Funnel config file.
// This function ensures that the config gets set up properly. Flag values override values in
// the provided config file.
func MergeConfigFileWithFlags(file string, flagConf *config.Config) (*config.Config, error) {
	// parse config file if it exists
	conf := config.DefaultConfig()
	err := config.ParseFile(file, conf)
	if err != nil {
		return conf, err
	}

	// file vals <- cli val
	proto.Merge(conf, flagConf)
	mergeDurations(conf, flagConf)
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
func TempConfigFile(c *config.Config, name string) (path string, cleanup func()) {
	tmpdir, err := os.MkdirTemp("", "")
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
