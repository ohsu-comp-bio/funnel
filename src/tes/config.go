package tes

import (
	"github.com/ghodss/yaml"
	os_servers "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
  log "tes/logger"
	pbr "tes/server/proto"
)

// Local describes configuration for the local scheduler.
type Local struct {
	NumWorkers int
}

// Openstack describes configuration for the openstack scheduler.
type Openstack struct {
	NumWorkers int
	KeyPair    string
	ConfigPath string
	Server     os_servers.CreateOpts
}

// Schedulers describes configuration for all schedulers.
type Schedulers struct {
	Local     Local
	Openstack Openstack
}

// Config describes configuration for TES.
type Config struct {
	pbr.ServerConfig
	// TODO move this to protobuf?
	Schedulers Schedulers
	DBPath     string
	HTTPPort   string
	RPCPort    string
	Scheduler  string
	ContentDir string
	WorkDir    string
}

// DefaultConfig returns configuration with simple defaults.
func DefaultConfig() Config {
	workDir := "tes-work-dir"
	return Config{
		ServerConfig: pbr.ServerConfig{
			ServerAddress: "localhost:9090",
		},
		DBPath:     path.Join(workDir, "tes_task.db"),
		HTTPPort:   "8000",
		RPCPort:    "9090",
		Scheduler:  "local",
		ContentDir: defaultContentDir(),
		WorkDir:    workDir,
		Schedulers: Schedulers{
			Local: Local{
				NumWorkers: 4,
			},
		},
	}
}

func defaultContentDir() string {
	// TODO this depends on having the entire repo available
	//      which prevents us from releasing a single binary.
	//      Not the worst, but maybe there's a good way to make it optional.
	// TODO handle error
	dir, _ := filepath.Abs(os.Args[0])
	return filepath.Join(dir, "..", "..", "share")
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

// LoadConfigOrExit tries to load the config from the given file.
// If the file cannot be loaded, os.Exit() is called.
func LoadConfigOrExit(relpath string, config interface{}) {
	var err error
	if relpath != "" {
		var abspath string
		abspath, err = filepath.Abs(relpath)
		if err != nil {
			log.Error("Failure reading config", "path", abspath, "error", err)
			os.Exit(1)
		}
		log.Info("Using config file", "path", abspath)
		err = ParseConfigFile(abspath, &config)
		if err != nil {
			log.Error("Failure reading config", "path", abspath, "error", err)
			os.Exit(1)
		}
	}
}
