package config

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/ohsu-comp-bio/funnel/logger"
	os_servers "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

// Config describes configuration for Funnel.
type Config struct {
	Storage   StorageConfig
	HostName  string
	Scheduler string
	Server    struct {
		Password string
	}
	Backends struct {
		Local    struct{}
		HTCondor struct {
			Template string
		}
		SLURM struct {
			Template string
		}
		PBS struct {
			Template string
		}
		GridEngine struct {
			Template string
		}
		OpenStack struct {
			KeyPair    string
			ConfigPath string
			Server     os_servers.CreateOpts
		}
		// Google Cloud Compute
		GCE struct {
			AccountFile string
			Project     string
			Zone        string
			Weights     struct {
				PreferQuickStartup float32
			}
			CacheTTL time.Duration
		}
	}
	Worker             Worker
	DBPath             string
	HTTPPort           string
	RPCPort            string
	WorkDir            string
	Logger             logger.Config
	MaxExecutorLogSize int
	ScheduleRate       time.Duration
	ScheduleChunk      int
	// How long to wait for a worker ping before marking it as dead
	WorkerPingTimeout time.Duration
	// How long to wait for worker initialization before marking it dead
	WorkerInitTimeout time.Duration
	DisableHTTPCache  bool
	ServiceName       string
}

// HTTPAddress returns the HTTP address based on HostName and HTTPPort
func (c Config) HTTPAddress() string {
	return "http://" + c.HostName + ":" + c.HTTPPort
}

// RPCAddress returns the RPC address based on HostName and RPCPort
func (c Config) RPCAddress() string {
	return c.HostName + ":" + c.RPCPort
}

// DefaultConfig returns configuration with simple defaults.
func DefaultConfig() Config {
	workDir := "funnel-work-dir"
	hostName := "localhost"
	rpcPort := "9090"
	cwd, _ := os.Getwd()
	c := Config{
		HostName:  hostName,
		DBPath:    path.Join(workDir, "funnel.db"),
		HTTPPort:  "8000",
		RPCPort:   rpcPort,
		WorkDir:   workDir,
		Logger:    logger.DefaultConfig(),
		Scheduler: "local",
		Storage: StorageConfig{
			Local: LocalStorage{
				AllowedDirs: []string{cwd},
			},
		},
		MaxExecutorLogSize: 10000,
		ScheduleRate:       time.Second,
		ScheduleChunk:      10,
		WorkerPingTimeout:  time.Minute,
		WorkerInitTimeout:  time.Minute * 5,
		Worker: Worker{
			ServerAddress: hostName + ":" + rpcPort,
			WorkDir:       workDir,
			Timeout:       -1,
			// TODO these get reset to zero when not found in yaml?
			UpdateRate:    time.Second * 5,
			LogUpdateRate: time.Second * 5,
			LogTailSize:   10000,
			Logger:        logger.DefaultConfig(),
			Resources:     Resources{},
			UpdateTimeout: time.Second,
			Metadata:      map[string]string{},
		},
		DisableHTTPCache: true,
		ServiceName:      "Funnel",
	}

	htcondorTemplate, _ := Asset("config/htcondor-template.txt")
	slurmTemplate, _ := Asset("config/slurm-template.txt")
	pbsTemplate, _ := Asset("config/pbs-template.txt")
	geTemplate, _ := Asset("config/gridengine-template.txt")

	c.Backends.HTCondor.Template = string(htcondorTemplate)
	c.Backends.SLURM.Template = string(slurmTemplate)
	c.Backends.PBS.Template = string(pbsTemplate)
	c.Backends.GridEngine.Template = string(geTemplate)
	c.Backends.GCE.CacheTTL = time.Minute
	c.Backends.GCE.Weights.PreferQuickStartup = 1.0
	return c
}

// Resources describes worker resource config.
type Resources struct {
	Cpus   uint32
	RamGb  float64 // nolint
	DiskGb float64
}

// Worker contains worker configuration.
type Worker struct {
	ID string
	// Address of the scheduler, e.g. "1.2.3.4:9090"
	ServerAddress string
	// Directory to write task files to
	WorkDir string
	// How long (seconds) to wait before tearing down an inactive worker
	// Default, -1, indicates to tear down the worker immediately after completing
	// its task
	Timeout time.Duration
	// How often the worker sends update requests to the server
	UpdateRate time.Duration
	// How often the worker sends task log updates
	LogUpdateRate time.Duration
	LogTailSize   int64
	Storage       StorageConfig
	Logger        logger.Config
	Resources     Resources
	// Timeout duration for UpdateWorker() and UpdateTaskLogs() RPC calls
	UpdateTimeout  time.Duration
	Metadata       map[string]string
	ServerPassword string
}

// WorkerInheritConfigVals is a utility to help ensure the Worker inherits the proper config values from the parent Config
func WorkerInheritConfigVals(c Config) Worker {
	if (c.HostName != "") && (c.RPCPort != "") {
		c.Worker.ServerAddress = c.HostName + ":" + c.RPCPort
	}
	c.Worker.Storage = c.Storage
	c.Worker.ServerPassword = c.Server.Password
	return c.Worker
}

// StorageConfig describes configuration for all storage types
type StorageConfig struct {
	Local LocalStorage
	S3    []S3Storage
	GS    []GSStorage
}

// LocalStorage describes the directories Funnel can read from and write to
type LocalStorage struct {
	AllowedDirs []string
}

// Valid validates the LocalStorage configuration
func (l LocalStorage) Valid() bool {
	return len(l.AllowedDirs) > 0
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

// S3Storage describes the directories Funnel can read from and write to
type S3Storage struct {
	Endpoint string
	Key      string
	Secret   string
}

// Valid validates the LocalStorage configuration
func (l S3Storage) Valid() bool {
	return l.Endpoint != "" && l.Key != "" && l.Secret != ""
}

// ToYaml formats the configuration into YAML and returns the bytes.
func (c Config) ToYaml() []byte {
	// TODO handle error
	yamlstr, _ := yaml.Marshal(c)
	return yamlstr
}

// ToYamlFile writes the configuration to a YAML file.
func (c Config) ToYamlFile(p string) {
	// TODO handle error
	ioutil.WriteFile(p, c.ToYaml(), 0600)
}

// ToYamlTempFile writes the configuration to a YAML temp. file.
func (c Config) ToYamlTempFile(name string) (string, func()) {
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

// Parse parses a YAML doc into the given Config instance.
func Parse(raw []byte, conf *Config) error {
	err := yaml.Unmarshal(raw, conf)
	if err != nil {
		return err
	}
	return nil
}

// ParseFile parses a Funnel config file, which is formatted in YAML,
// and returns a Config struct.
func ParseFile(relpath string, conf *Config) error {
	if relpath == "" {
		return nil
	}

	// Try to get absolute path. If it fails, fall back to relative path.
	path, abserr := filepath.Abs(relpath)
	if abserr != nil {
		path = relpath
	}

	// Read file
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("can't read config: %s", err)
	}

	// Parse file
	perr := Parse(source, conf)
	if perr != nil {
		return fmt.Errorf("can't read config: %s", perr)
	}
	return nil
}
