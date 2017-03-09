package config

import (
	"github.com/ghodss/yaml"
	os_servers "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	log "tes/logger"
	pbr "tes/server/proto"
	"time"
)

// Weights describes the scheduler score weights.
// All fields should be float32 type.
type Weights map[string]float32

// StorageConfig describes configuration for all storage types
type StorageConfig struct {
	Local LocalStorage
	S3    S3Storage
	GS    GSStorage
}

// LocalStorage describes the directories TES can read from and write to
type LocalStorage struct {
	AllowedDirs []string
}

// GSStorage describes configuration for the Google Cloud storage backend.
type GSStorage struct {
	AccountFile string
	FromEnv     bool
}

// Valid validates the GSStorage configuration.
func (g GSStorage) Valid() bool {
	return g.FromEnv || g.AccountFile != ""
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
	Weights Weights
}

// OpenstackScheduler describes configuration for the openstack scheduler.
type OpenstackScheduler struct {
	NumWorkers int
	KeyPair    string
	ConfigPath string
	Server     os_servers.CreateOpts
}

// GCEScheduler describes configuration for the Google Cloud scheduler.
type GCEScheduler struct {
	AccountFile string
	Project     string
	Zone        string
	Templates   []string
	Weights     Weights
}

// Schedulers describes configuration for all schedulers.
type Schedulers struct {
	Local     LocalScheduler
	Dumblocal LocalScheduler
	Condor    LocalScheduler
	Openstack OpenstackScheduler
	GCE       GCEScheduler
}

// Config describes configuration for TES.
type Config struct {
	Storage       []*StorageConfig
	HostName        string
	Scheduler     string
	Schedulers    Schedulers
	Worker        Worker
	DBPath        string
	HTTPPort      string
	RPCPort       string
	ContentDir    string
	WorkDir       string
	LogLevel      string
	MaxJobLogSize int
	ScheduleRate  time.Duration
	ScheduleChunk int
	// How long to wait for a worker ping before marking it as dead
	WorkerPingTimeout time.Duration
	// How long to wait for worker initialization before marking it dead
	WorkerInitTimeout time.Duration
}

// DefaultConfig returns configuration with simple defaults.
func DefaultConfig() Config {
	workDir := "tes-work-dir"
	hostName := "localhost"
	rpcPort := "9090"
	c := Config{
		HostName:   hostName,
		DBPath:     path.Join(workDir, "tes_task.db"),
		HTTPPort:   "8000",
		RPCPort:    rpcPort,
		ContentDir: defaultContentDir(),
		WorkDir:    workDir,
		LogLevel:   "debug",
		Scheduler:  "local",
		Schedulers: Schedulers{
			Local: LocalScheduler{},
			GCE: GCEScheduler{
				Weights: Weights{
					"startup time": 1.0,
				},
			},
		},
		MaxJobLogSize:     10000,
		ScheduleRate:      time.Second,
		ScheduleChunk:     10,
		WorkerPingTimeout: time.Minute,
		WorkerInitTimeout: time.Minute * 5,
	}
	c.Worker = WorkerDefaultConfig(c)
	return c
}

// Worker contains worker configuration.
type Worker struct {
	ID string
	// Address of the scheduler, e.g. "1.2.3.4:9090"
	ServerAddress string
	// Directory to write job files to
	WorkDir string
	// How long (seconds) to wait before tearing down an inactive worker
	// Default, -1, indicates to tear down the worker immediately after completing
	// its job
	Timeout time.Duration
	// How often the worker sends update requests to the server
	UpdateRate time.Duration
	// How often the worker sends job log updates
	LogUpdateRate time.Duration
	TrackerRate   time.Duration
	LogTailSize   int64
	Storage       []*StorageConfig
	LogPath       string
	LogLevel      string
	Resources     *pbr.Resources
	// Timeout duration for UpdateWorker() and UpdateJobLogs() RPC calls
	UpdateTimeout time.Duration
	Metadata      map[string]string
}

// WorkerDefaultConfig returns simple, default worker configuration.
func WorkerDefaultConfig(c Config) Worker {
	return Worker{
		ServerAddress: c.HostName + ":" + c.RPCPort,
		WorkDir:       c.WorkDir,
		Timeout:       -1,
		// TODO these get reset to zero when not found in yaml?
		UpdateRate:    time.Second * 5,
		LogUpdateRate: time.Second * 5,
		TrackerRate:   time.Second * 5,
		LogTailSize:   10000,
		Storage:       c.Storage,
		LogLevel:      "debug",
		UpdateTimeout: time.Second,
	}
}

// WorkerInheritConfigVals is a utility to help ensure the Worker inherits the proper config values from the parent Config
func WorkerInheritConfigVals(c Config) Worker {
	c.Worker.ServerAddress = c.HostName + ":" + c.RPCPort
	c.Worker.WorkDir = c.WorkDir
	c.Worker.Storage = c.Storage
	return c.Worker
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
