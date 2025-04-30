package util

import (
	"os"
	"path/filepath"
	"reflect"
	"time"

	"dario.cat/mergo"
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
	if base == nil || override == nil {
		return
	}
	bv := reflect.ValueOf(base).Elem()
	ov := reflect.ValueOf(override).Elem()
	for i := 0; i < bv.NumField(); i++ {
		bf := bv.Field(i)
		of := ov.Field(i)
		// Only work with pointer fields
		if of.Kind() == reflect.Ptr && !of.IsNil() {
			switch bf.Type().String() {
			case "*durationpb.Duration":
				overrideDur := of.Interface().(*durationpb.Duration)
				if overrideDur != nil && overrideDur.AsDuration() != time.Duration(0) {
					// Set base field to the override value
					bf.Set(reflect.ValueOf(overrideDur))
				}
			default:
				// Recurse into nested messages if both base and override are proto messages
				if bf.Type().Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) &&
					of.Type().Implements(reflect.TypeOf((*proto.Message)(nil)).Elem()) {

					if bf.IsNil() {
						// If base field is nil but override is not, make a new instance
						newField := reflect.New(bf.Type().Elem())
						bf.Set(newField)
					}

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
	err = mergo.MergeWithOverwrite(conf, flagConf)
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
