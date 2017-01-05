package tes

import (
	"github.com/ghodss/yaml"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"tes/scheduler/condor"
	"tes/scheduler/openstack"
	pbr "tes/server/proto"
)

type Config struct {
	pbr.ServerConfig
	Schedulers struct {
		Openstack openstack.Config
		Condor    condor.Config
	}
}

// ParseConfigFile parses a TES config file, which is formatted in YAML,
// and returns a ServerConfig struct.
func ParseConfigFile(path string, doc interface{}) error {
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(source, &doc)
	if err != nil {
		return err
	}
	return nil
}

func LoadConfigOrExit(relpath string, config interface{}) {
	var err error
	if relpath != "" {
		var abspath string
		abspath, err = filepath.Abs(relpath)
		if err != nil {
			log.Printf("Failure reading config: %s", abspath)
			os.Exit(1)
		}
		log.Printf("Using config file: %s", abspath)
		err = ParseConfigFile(abspath, &config)
		if err != nil {
			log.Printf("Failure reading config: %s", abspath)
			os.Exit(1)
		}
	}
}
