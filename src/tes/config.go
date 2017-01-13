package tes

import (
	"github.com/ghodss/yaml"
	os_servers "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	pbr "tes/server/proto"
)

type Local struct {
	NumWorkers int
}

type Openstack struct {
	NumWorkers int
	KeyPair    string
	ConfigPath string
	Server     os_servers.CreateOpts
}

type Schedulers struct {
	Local     Local
	Openstack Openstack
}

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

func LoadConfigOrExit(relpath string, config interface{}) {
	var err error
	if relpath != "" {
		var abspath string
		abspath, err = filepath.Abs(relpath)
		if err != nil {
			log.Printf("Failure reading config: %s", abspath)
			log.Println(err)
			os.Exit(1)
		}
		log.Printf("Using config file: %s", abspath)
		err = ParseConfigFile(abspath, &config)
		if err != nil {
			log.Printf("Failure reading config: %s", abspath)
			log.Println(err)
			os.Exit(1)
		}
	}
}
