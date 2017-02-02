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
	Dumblocal Local
	Condor    Local
	Openstack Openstack
}

// Config describes configuration for TES.
type Config struct {
	pbr.ServerConfig
	Scheduler  string
	Schedulers Schedulers
	Worker     Worker
	DBPath     string
	HTTPPort   string
	RPCPort    string
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
		ContentDir: defaultContentDir(),
		WorkDir:    workDir,
		Scheduler:  "Local",
		Schedulers: Schedulers{
			Local: Local{
				NumWorkers: 4,
			},
		},
		Worker: WorkerDefaultConfig(),
	}
}

// Worker contains worker configuration.
type Worker struct {
	ID string	
	Slots int
	// Address of the scheduler, e.g. "1.2.3.4:9090"
	ServerAddress string
	// Directory to write job files to
	WorkDir string
	Timeout int
	// how often (milliseconds) the worker polls for cancellation requests
	StatusPollRate int
	// how often (milliseconds) the worker sends log updates
	LogUpdateRate int
	Storage       []*pbr.StorageConfig
	LogPath       string
}

// WorkerDefaultConfig returns simple, default worker configuration.
func WorkerDefaultConfig() Worker {
	return Worker{
		Slots:          1,
		ServerAddress:  "localhost:9090",
		WorkDir:        "tes-work-dir",
		Timeout:        -1,
		StatusPollRate: 5000,
		LogUpdateRate:  5000,
		LogPath:        "tes-worker-log",
	}
}

// ToYaml formats the configuration into YAML and returns the bytes.
func (c Worker) ToYaml() []byte {
	// TODO handle error
	yamlstr, _ := yaml.Marshal(c)
	return yamlstr
}

// ToYamlFile writes the configuration to a YAML file.
func (c Worker) ToYamlFile(p string) {
	// TODO handle error
	ioutil.WriteFile(p, c.ToYaml(), 0600)
}

// ToYamlTempFile writes the configuration to a YAML temp. file.
func (c Worker) ToYamlTempFile(name string) (string, func()) {
	// I'm creating a temp. directory instead of a temp. file so that
	// the file can have an expected name. This is helpful for the HTCondor scheduler.
	tmpdir, _ := ioutil.TempDir("", "")

	cleanup := func() {
		os.RemoveAll(tmpdir)
	}

	p := filepath.Join(tmpdir, name)
	c.ToYamlFile(p)
	return p, cleanup
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
