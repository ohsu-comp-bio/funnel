package config

import (
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/ohsu-comp-bio/funnel/logger"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
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
		Batch AWSBatch
	}
	Scheduler Scheduler
	Worker    Worker
}

// EnsureServerProperties ensures that the server address and server password
// is consistent between the worker, node, and server.
func EnsureServerProperties(conf Config) Config {
	conf.Worker.EventWriters.RPC.ServerAddress = conf.Server.RPCAddress()
	conf.Worker.TaskReaders.RPC.ServerAddress = conf.Server.RPCAddress()
	conf.Scheduler.Node.ServerAddress = conf.Server.RPCAddress()

	conf.Worker.EventWriters.RPC.ServerPassword = conf.Server.Password
	conf.Worker.TaskReaders.RPC.ServerPassword = conf.Server.Password
	conf.Scheduler.Node.ServerPassword = conf.Server.Password
	return conf
}

// DefaultConfig returns configuration with simple defaults.
func DefaultConfig() Config {
	cwd, _ := os.Getwd()
	workDir := path.Join(cwd, "funnel-work-dir")

	server := Server{
		HostName:         "localhost",
		HTTPPort:         "8000",
		RPCPort:          "9090",
		ServiceName:      "Funnel",
		DisableHTTPCache: true,
		Logger:           logger.DefaultConfig(),
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

	rpc := RPC{
		ServerAddress:  server.RPCAddress(),
		ServerPassword: server.Password,
		Timeout:        time.Second,
	}
	dynamo := DynamoDB{
		TableBasename: "funnel",
	}
	elastic := Elastic{
		URL:         "http://localhost:9200",
		IndexPrefix: "funnel",
	}
	mongo := MongoDB{
		Addrs:    []string{"localhost"},
		Database: "funnel",
	}

	c.Server.Database = "boltdb"
	c.Server.Databases.BoltDB.Path = path.Join(workDir, "funnel.db")
	c.Server.Databases.DynamoDB = dynamo
	c.Server.Databases.Elastic = elastic
	c.Server.Databases.MongoDB = mongo

	c.Worker.TaskReader = "rpc"
	c.Worker.TaskReaders.RPC = rpc
	c.Worker.TaskReaders.DynamoDB = dynamo
	c.Worker.TaskReaders.Elastic = elastic
	c.Worker.TaskReaders.MongoDB = mongo

	c.Worker.ActiveEventWriters = []string{"rpc", "log"}
	c.Worker.EventWriters.RPC = rpc
	c.Worker.EventWriters.DynamoDB = dynamo
	c.Worker.EventWriters.Elastic = elastic
	c.Worker.EventWriters.MongoDB = mongo
	c.Worker.EventWriters.Kafka.Topic = "funnel"

	c.Worker.Storage.AmazonS3.MaxRetries = 10

	htcondorTemplate, _ := Asset("config/htcondor-template.txt")
	slurmTemplate, _ := Asset("config/slurm-template.txt")
	pbsTemplate, _ := Asset("config/pbs-template.txt")
	geTemplate, _ := Asset("config/gridengine-template.txt")

	c.Backends.HTCondor.Template = string(htcondorTemplate)
	c.Backends.SLURM.Template = string(slurmTemplate)
	c.Backends.PBS.Template = string(pbsTemplate)
	c.Backends.GridEngine.Template = string(geTemplate)

	c.Backends.Batch.JobDefinition = "funnel-job-def"
	c.Backends.Batch.JobQueue = "funnel-job-queue"

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
		Elastic  Elastic
		MongoDB  MongoDB
	}
	DisableHTTPCache bool
	Logger           logger.Config
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
	// Timeout duration for PutNode() gRPC calls
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
		RPC      RPC
		DynamoDB DynamoDB
		Elastic  Elastic
		MongoDB  MongoDB
	}
	ActiveEventWriters []string
	EventWriters       struct {
		RPC      RPC
		DynamoDB DynamoDB
		Elastic  Elastic
		MongoDB  MongoDB
		Kafka    Kafka
	}
}

// RPC configures access to the Funnel RPC server.
type RPC struct {
	// RPC address of the Funnel server
	ServerAddress string
	// Password for basic auth. with the server APIs.
	ServerPassword string
	// Timeout duration for gRPC calls
	Timeout time.Duration
}

// MongoDB configures access to an MongoDB database.
type MongoDB struct {
	// Addrs holds the addresses for the seed servers.
	Addrs []string
	// Database is the database name used within MongoDB to store funnel data.
	Database string
	Username string
	Password string
}

// Elastic configures access to an Elasticsearch database.
type Elastic struct {
	IndexPrefix string
	URL         string
}

// Kafka configure access to a Kafka topic for task event reading/writing.
type Kafka struct {
	Servers []string
	Topic   string
}

// AWSConfig describes the configuration for creating AWS Session instances
type AWSConfig struct {
	// An optional endpoint URL (hostname only or fully qualified URI)
	// that overrides the default generated endpoint for a client.
	Endpoint string
	// The region to send requests to. This parameter is required and must
	// be configured on a per-client basis for all AWS services in funnel with the
	// exception of S3Storage.
	Region string
	// The maximum number of times that a request will be retried for failures.
	// By default defers the max retry setting to the service
	// specific configuration. Set to a value > 0 to override.
	MaxRetries int
	// If both the key and secret are empty AWS credentials will be read from
	// the environment.
	Key    string
	Secret string
}

// AWSBatch describes the configuration for the AWS Batch compute backend.
type AWSBatch struct {
	// JobDefinition can be either a name or the Amazon Resource Name (ARN).
	JobDefinition string
	// JobQueue can be either a name or the Amazon Resource Name (ARN).
	JobQueue string
	AWS      AWSConfig
}

// DynamoDB describes the configuration for Amazon DynamoDB backed processes
// such as the event writer and server.
type DynamoDB struct {
	AWS           AWSConfig
	TableBasename string
}

// StorageConfig describes configuration for all storage types
type StorageConfig struct {
	Local     LocalStorage
	AmazonS3  AmazonS3Storage
	GenericS3 []GenericS3Storage
	GS        GSStorage
	Swift     SwiftStorage
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
	Disabled bool
	// If no account file is provided then Funnel will try to use Google Application
	// Default Credentials to authorize and authenticate the client.
	AccountFile string
}

// Valid validates the GSStorage configuration.
func (g GSStorage) Valid() bool {
	return !g.Disabled
}

// AmazonS3Storage describes the configuration for the Amazon S3 storage backend.
type AmazonS3Storage struct {
	Disabled bool
	AWSConfig
}

// Valid validates the AmazonS3Storage configuration
func (s AmazonS3Storage) Valid() bool {
	creds := (s.Key != "" && s.Secret != "") || (s.Key == "" && s.Secret == "")
	return !s.Disabled && creds
}

// GenericS3Storage describes the configuration for the Generic S3 storage backend.
type GenericS3Storage struct {
	Disabled bool
	Endpoint string
	Key      string
	Secret   string
}

// Valid validates the S3Storage configuration
func (s GenericS3Storage) Valid() bool {
	return !s.Disabled && s.Key != "" && s.Secret != "" && s.Endpoint != ""
}

// SwiftStorage configures the OpenStack Swift object storage backend.
type SwiftStorage struct {
	Disabled   bool
	UserName   string
	Password   string
	AuthURL    string
	TenantName string
	TenantID   string
	RegionName string
	// Size of chunks to use for large object creation.
	// Defaults to 500 MB if not set or set below 10 MB.
	// The max number of chunks for a single object is 1000.
	ChunkSizeBytes int64
}

// Valid validates the SwiftStorage configuration.
func (s SwiftStorage) Valid() bool {
	user := s.UserName != "" || os.Getenv("OS_USERNAME") != ""
	password := s.Password != "" || os.Getenv("OS_PASSWORD") != ""
	authURL := s.AuthURL != "" || os.Getenv("OS_AUTH_URL") != ""
	tenantName := s.TenantName != "" || os.Getenv("OS_TENANT_NAME") != "" || os.Getenv("OS_PROJECT_NAME") != ""
	tenantID := s.TenantID != "" || os.Getenv("OS_TENANT_ID") != "" || os.Getenv("OS_PROJECT_ID") != ""
	region := s.RegionName != "" || os.Getenv("OS_REGION_NAME") != ""

	valid := user && password && authURL && tenantName && tenantID && region

	return !s.Disabled && valid
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
		return fmt.Errorf("Failure reading config at path %s: %s", path, err)
	}

	// Parse file
	perr := Parse(source, conf)
	if perr != nil {
		return fmt.Errorf("Failure reading config at path %s: %s", path, err)
	}
	return nil
}
