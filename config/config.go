package config

import (
	"github.com/ghodss/yaml"
	"github.com/ohsu-comp-bio/funnel/logger"
	os_servers "github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Config describes configuration for Funnel.
type Config struct {
	Server Server
	// the active compute backend
	Backend  string
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
	Scheduler Scheduler
	Worker    Worker
}

// EnsureServerProperties ensures that the server address and server password
// is consistent between the worker, node, and server.
//
// Precedence:
// Server > Node > TaskReader > EventWriter
func EnsureServerProperties(conf Config) Config {
	conf = ensureServerAddress(conf)
	conf = ensureServerPassword(conf)
	return conf
}

func ensureServerAddress(conf Config) Config {
	defaults := DefaultConfig()
	empty := Config{}

	if conf.Server.RPCAddress() != defaults.Server.RPCAddress() && conf.Server.RPCAddress() != empty.Server.RPCAddress() {
		conf.Worker.EventWriters.RPC.ServerAddress = conf.Server.RPCAddress()
		conf.Worker.TaskReaders.RPC.ServerAddress = conf.Server.RPCAddress()
		conf.Scheduler.Node.ServerAddress = conf.Server.RPCAddress()
	}

	if conf.Scheduler.Node.ServerAddress != conf.Server.RPCAddress() && conf.Scheduler.Node.ServerAddress != empty.Scheduler.Node.ServerAddress {
		parts := strings.Split(conf.Scheduler.Node.ServerAddress, ":")
		conf.Server.HostName = parts[0]
		conf.Server.RPCPort = parts[1]
		conf.Worker.TaskReaders.RPC.ServerAddress = conf.Scheduler.Node.ServerAddress
		conf.Worker.EventWriters.RPC.ServerAddress = conf.Scheduler.Node.ServerAddress
	}

	if conf.Worker.TaskReaders.RPC.ServerAddress != conf.Server.RPCAddress() && conf.Worker.TaskReaders.RPC.ServerAddress != empty.Worker.TaskReaders.RPC.ServerAddress {
		parts := strings.Split(conf.Worker.TaskReaders.RPC.ServerAddress, ":")
		conf.Server.HostName = parts[0]
		conf.Server.RPCPort = parts[1]
		conf.Scheduler.Node.ServerAddress = conf.Worker.TaskReaders.RPC.ServerAddress
		conf.Worker.EventWriters.RPC.ServerAddress = conf.Worker.TaskReaders.RPC.ServerAddress
	}

	if conf.Worker.EventWriters.RPC.ServerAddress != conf.Server.RPCAddress() && conf.Worker.EventWriters.RPC.ServerAddress != empty.Worker.EventWriters.RPC.ServerAddress {
		parts := strings.Split(conf.Worker.EventWriters.RPC.ServerAddress, ":")
		conf.Server.HostName = parts[0]
		conf.Server.RPCPort = parts[1]
		conf.Scheduler.Node.ServerAddress = conf.Worker.EventWriters.RPC.ServerAddress
		conf.Worker.TaskReaders.RPC.ServerAddress = conf.Worker.EventWriters.RPC.ServerAddress
	}

	return conf
}

func ensureServerPassword(conf Config) Config {
	defaults := DefaultConfig()
	empty := Config{}

	if conf.Server.Password != defaults.Server.Password && conf.Server.Password != empty.Server.Password {
		conf.Worker.EventWriters.RPC.ServerPassword = conf.Server.Password
		conf.Worker.TaskReaders.RPC.ServerPassword = conf.Server.Password
		conf.Scheduler.Node.ServerPassword = conf.Server.Password
	}

	if conf.Scheduler.Node.ServerPassword != conf.Server.Password && conf.Scheduler.Node.ServerPassword != empty.Scheduler.Node.ServerPassword {
		conf.Server.Password = conf.Scheduler.Node.ServerPassword
		conf.Worker.EventWriters.RPC.ServerPassword = conf.Scheduler.Node.ServerPassword
		conf.Worker.TaskReaders.RPC.ServerPassword = conf.Scheduler.Node.ServerPassword
	}

	if conf.Worker.TaskReaders.RPC.ServerPassword != conf.Server.Password && conf.Worker.TaskReaders.RPC.ServerPassword != empty.Worker.TaskReaders.RPC.ServerPassword {
		conf.Server.Password = conf.Worker.TaskReaders.RPC.ServerPassword
		conf.Scheduler.Node.ServerPassword = conf.Worker.TaskReaders.RPC.ServerPassword
	}

	if conf.Worker.EventWriters.RPC.ServerPassword != conf.Server.Password && conf.Worker.EventWriters.RPC.ServerPassword != empty.Worker.EventWriters.RPC.ServerPassword {
		conf.Server.Password = conf.Worker.EventWriters.RPC.ServerPassword
		conf.Scheduler.Node.ServerPassword = conf.Worker.EventWriters.RPC.ServerPassword
	}

	return conf
}

// DefaultConfig returns configuration with simple defaults.
func DefaultConfig() Config {
	cwd, _ := os.Getwd()
	workDir := path.Join(cwd, "funnel-work-dir")

	server := Server{
		HostName:           "localhost",
		HTTPPort:           "8000",
		RPCPort:            "9090",
		ServiceName:        "Funnel",
		MaxExecutorLogSize: 10000,
		DisableHTTPCache:   true,
		Logger:             logger.DefaultConfig(),
	}

	c := Config{
		Server:  server,
		Backend: "local",
		Scheduler: Scheduler{
			ScheduleRate:    time.Second,
			ScheduleChunk:   10,
			NodePingTimeout: time.Minute,
			NodeInitTimeout: time.Minute * 5,
			NodeDeadTimeout: time.Minute * 5,
			Node: Node{
				ServerAddress:  server.RPCAddress(),
				ServerPassword: server.Password,
				WorkDir:        workDir,
				Timeout:        -1,
				UpdateRate:     time.Second * 5,
				UpdateTimeout:  time.Second,
				Metadata:       map[string]string{},
				Logger:         logger.DefaultConfig(),
			},
			Logger: logger.DefaultConfig(),
		},
		Worker: Worker{
			WorkDir: workDir,
			Storage: StorageConfig{
				Local: LocalStorage{
					AllowedDirs: []string{cwd},
				},
			},
			UpdateRate: time.Second * 5,
			BufferSize: 10000,
			Logger:     logger.DefaultConfig(),
		},
	}

	c.Server.Database = "boltdb"
	c.Server.Databases.BoltDB.Path = path.Join(workDir, "funnel.db")
	c.Server.Databases.DynamoDB.TableBasename = "funnel"

	c.Worker.TaskReader = "rpc"
	c.Worker.TaskReaders.RPC.ServerAddress = server.RPCAddress()
	c.Worker.TaskReaders.RPC.ServerPassword = server.Password
	c.Worker.TaskReaders.DynamoDB.TableBasename = "funnel"

	c.Worker.ActiveEventWriters = []string{"rpc", "log"}
	c.Worker.EventWriters.RPC.ServerAddress = server.RPCAddress()
	c.Worker.EventWriters.RPC.ServerPassword = server.Password
	c.Worker.EventWriters.RPC.UpdateTimeout = time.Second
	c.Worker.EventWriters.DynamoDB.TableBasename = "funnel"

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

// Server describes configuration for the server.
type Server struct {
	ServiceName string
	HostName    string
	HTTPPort    string
	RPCPort     string
	Password    string
	Database    string
	Databases   struct {
		BoltDB struct {
			Path string
		}
		DynamoDB DynamoDB
	}
	DisableHTTPCache   bool
	MaxExecutorLogSize int
	Logger             logger.Config
}

// HTTPAddress returns the HTTP address based on HostName and HTTPPort
func (c Server) HTTPAddress() string {
	if c.HostName != "" && c.HTTPPort != "" {
		return "http://" + c.HostName + ":" + c.HTTPPort
	}
	return ""
}

// RPCAddress returns the RPC address based on HostName and RPCPort
func (c *Server) RPCAddress() string {
	if c.HostName != "" && c.RPCPort != "" {
		return c.HostName + ":" + c.RPCPort
	}
	return ""
}

// Scheduler contains funnel's basic scheduler configuration.
type Scheduler struct {
	// How often to run a scheduler iteration.
	ScheduleRate time.Duration
	// How many tasks to schedule in one iteration.
	ScheduleChunk int
	// How long to wait for a node ping before marking it as dead
	NodePingTimeout time.Duration
	// How long to wait for node initialization before marking it dead
	NodeInitTimeout time.Duration
	// How long to wait before deleting a dead node from the DB.
	NodeDeadTimeout time.Duration
	// Node configuration
	Node Node
	// Logger configuration
	Logger logger.Config
}

// Node contains the configuration for a node. Nodes track available resources
// for funnel's basic scheduler.
type Node struct {
	ID      string
	WorkDir string
	// A Node will automatically try to detect what resources are available to it.
	// Defining Resources in the Node configuration overrides this behavior.
	Resources struct {
		Cpus   uint32
		RamGb  float64 // nolint
		DiskGb float64
	}
	// If the node has been idle for longer than the timeout, it will shut down.
	// -1 means there is no timeout. 0 means timeout immediately after the first task.
	Timeout time.Duration
	// How often the node sends update requests to the server.
	UpdateRate time.Duration
	// Timeout duration for UpdateNode() gRPC calls
	UpdateTimeout time.Duration
	Metadata      map[string]string
	// RPC address of the Funnel server
	ServerAddress string
	// Password for basic auth. with the server APIs.
	ServerPassword string
	Logger         logger.Config
}

// Worker contains worker configuration.
type Worker struct {
	// Directory to write task files to
	WorkDir string
	// How often the worker sends task log updates
	UpdateRate time.Duration
	// Max bytes to store in-memory between updates
	BufferSize  int64
	Storage     StorageConfig
	Logger      logger.Config
	TaskReader  string
	TaskReaders struct {
		RPC struct {
			// RPC address of the Funnel server
			ServerAddress string
			// Password for basic auth. with the server APIs.
			ServerPassword string
		}
		DynamoDB DynamoDB
	}
	ActiveEventWriters []string
	EventWriters       struct {
		RPC struct {
			// RPC address of the Funnel server
			ServerAddress string
			// Password for basic auth. with the server APIs.
			ServerPassword string
			// Timeout duration for gRPC calls
			UpdateTimeout time.Duration
		}
		DynamoDB DynamoDB
	}
}

// DynamoDB describes the configuration for Amazon DynamoDB backed processes
// such as the event writer and server.
type DynamoDB struct {
	Region        string
	Key           string
	Secret        string
	TableBasename string
}

// StorageConfig describes configuration for all storage types
type StorageConfig struct {
	Local LocalStorage
	S3    S3Storage
	GS    []GSStorage
	Swift SwiftStorage
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
	Key     string
	Secret  string
	FromEnv bool
}

// Valid validates the LocalStorage configuration
func (l S3Storage) Valid() bool {
	return (l.Key != "" && l.Secret != "") || l.FromEnv
}

// SwiftStorage configures the OpenStack Swift object storage backend.
type SwiftStorage struct {
	UserName   string
	Password   string
	AuthURL    string
	TenantName string
	TenantID   string
	RegionName string
}

// Valid validates the SwiftStorage configuration.
func (s SwiftStorage) Valid() bool {
	return s.UserName != "" && s.Password != "" && s.AuthURL != "" &&
		s.TenantName != "" && s.TenantID != "" && s.RegionName != ""
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
		logger.Error("Failure reading config", "path", path, "error", err)
		return err
	}

	// Parse file
	perr := Parse(source, conf)
	if perr != nil {
		logger.Error("Failure reading config", "path", path, "error", perr)
		return perr
	}
	return nil
}
