package config

import (
	"github.com/ghodss/yaml"
	os_servers "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	log "tes/logger"
	"time"
)

// StorageConfig describes configuration for all storage types
type StorageConfig struct {
	Local LocalStorage
	S3    S3Storage
}

// LocalStorage describes the directories TES can read from and write to
type LocalStorage struct {
	AllowedDirs []string
}

// Valid validates the LocalStorage configuration
func (l LocalStorage) Valid() bool {
	return len(l.AllowedDirs) > 0
}

// S3Storage describes the directories TES can read from and write to
type S3Storage struct {
	Endpoint string
	Key      string
	Secret   string
}

// Valid validates the LocalStorage configuration
func (l S3Storage) Valid() bool {
	return l.Endpoint != "" && l.Key != "" && l.Secret != ""
}

// LocalScheduler describes configuration for the local scheduler.
type LocalScheduler struct {
	NumWorkers int
}

// OpenstackScheduler describes configuration for the openstack scheduler.
type OpenstackScheduler struct {
	NumWorkers int
	KeyPair    string
	ConfigPath string
	Server     os_servers.CreateOpts
}

// Schedulers describes configuration for all schedulers.
type Schedulers struct {
	Local     LocalScheduler
	Dumblocal LocalScheduler
	Condor    LocalScheduler
	Openstack OpenstackScheduler
}

// Config describes configuration for TES.
type Config struct {
	Storage       []*StorageConfig
	ServerAddress string
	Scheduler     string
	Schedulers    Schedulers
	Worker        Worker
	DBPath        string
	HTTPPort      string
	RPCPort       string
	ContentDir    string
	WorkDir       string
	LogLevel      string
}

// DefaultConfig returns configuration with simple defaults.
func DefaultConfig() Config {
	workDir := "tes-work-dir"
	return Config{
		ServerAddress: "localhost:9090",
		DBPath:        path.Join(workDir, "tes_task.db"),
		HTTPPort:      "8000",
		RPCPort:       "9090",
		ContentDir:    defaultContentDir(),
		WorkDir:       workDir,
		LogLevel:      "debug",
		Scheduler:     "local",
		Schedulers: Schedulers{
			Local: LocalScheduler{
				NumWorkers: 4,
			},
		},
		Worker: WorkerDefaultConfig(),
	}
}

// Worker contains worker configuration.
type Worker struct {
	ID string
	// How many jobs can a worker accept at a time
	Slots int
	// Address of the scheduler, e.g. "1.2.3.4:9090"
	ServerAddress string
	// Directory to write job files to
	WorkDir string
	// How long (seconds) to wait before tearing down an inactive worker
	// Default, -1, indicates to tear down the worker immediately after completing
	// its job
	Timeout time.Duration
	// How often (milliseconds) the worker polls for cancellation requests
	StatusPollRate time.Duration
	// How often (milliseconds) the worker sends log updates
	LogUpdateRate time.Duration
	// How often (milliseconds) the scheduler polls for new jobs
	NewJobPollRate time.Duration
	Storage        []*StorageConfig
	LogPath        string
	LogLevel       string
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
		NewJobPollRate: 5000,
		LogLevel:       "debug",
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
// and returns a Config struct.
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
