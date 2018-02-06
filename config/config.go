package config

import (
	"os"
	"time"

	"github.com/ohsu-comp-bio/funnel/logger"
)

// Config describes configuration for Funnel.
type Config struct {
	// component selectors
	EventWriters []string
	Database     string
	Compute      string
	// funnel components
	Server    Server
	Scheduler Scheduler
	Node      Node
	Worker    Worker
	Logger    logger.Config
	// databases / event handlers
	BoltDB    BoltDB
	DynamoDB  DynamoDB
	Elastic   Elastic
	MongoDB   MongoDB
	Kafka     Kafka
	PubSub    PubSub
	Datastore Datastore
	// compute
	HTCondor   HPCBackend
	Slurm      HPCBackend
	PBS        HPCBackend
	GridEngine HPCBackend
	AWSBatch   AWSBatch
	// storage
	LocalStorage  LocalStorage
	AmazonS3      AmazonS3Storage
	GenericS3     []GenericS3Storage
	GoogleStorage GSStorage
	Swift         SwiftStorage
	HTTPStorage   HTTPStorage
}

// Server describes configuration for the server.
type Server struct {
	ServiceName      string
	HostName         string
	HTTPPort         string
	RPCPort          string
	Password         string
	DisableHTTPCache bool
	// The timeout to use for making RPC client connections in nanoseconds
	// This timeout is Only enforced when used in conjunction with the
	// grpc.WithBlock dial option.
	RPCClientTimeout time.Duration
	// The maximum number of times that a request will be retried for failures.
	// Time between retries follows an exponential backoff starting at 5 seconds
	// up to 1 minute
	RPCClientMaxRetries uint
}

// HTTPAddress returns the HTTP address based on HostName and HTTPPort
func (c Server) HTTPAddress() string {
	http := ""
	if c.HostName != "" {
		http = "http://" + c.HostName
	}
	if c.HTTPPort != "" {
		http = http + ":" + c.HTTPPort
	}
	return http
}

// RPCAddress returns the RPC address based on HostName and RPCPort
func (c *Server) RPCAddress() string {
	rpc := c.HostName
	if c.RPCPort != "" {
		rpc = rpc + ":" + c.RPCPort
	}
	return rpc
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
}

// Node contains the configuration for a node. Nodes track available resources
// for funnel's basic scheduler.
type Node struct {
	ID string
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
	Metadata   map[string]string
}

// Worker contains worker configuration.
type Worker struct {
	// Directory to write task files to
	WorkDir string
	// How often the worker sends task log updates
	UpdateRate time.Duration
	// Max bytes to store in-memory between updates
	BufferSize int64
	// Normally the worker cleans up its working directory after executing.
	// This option disables that behavior.
	LeaveWorkDir bool
}

// HPCBackend describes the configuration for a HPC scheduler backend such as
// HTCondor or Slurm.
type HPCBackend struct {
	Template string
}

// BoltDB describes the configuration for the BoltDB embedded database.
type BoltDB struct {
	Path string
}

// MongoDB configures access to an MongoDB database.
type MongoDB struct {
	// Addrs holds the addresses for the seed servers.
	Addrs []string
	// Database is the database name used within MongoDB to store funnel data.
	Database string
	// Timeout is the amount of time to wait for a server to respond when
	// first connecting and on follow up operations in the session. If
	// timeout is zero, the call may block forever waiting for a connection
	// to be established.
	Timeout time.Duration
	// Username and Password inform the credentials for the initial authentication
	// done on the database defined by the Database field.
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

// PubSub configures access to Google Cloud Pub/Sub for task event reading/writing.
type PubSub struct {
	Topic   string
	Project string
	// If no account file is provided then Funnel will try to use Google Application
	// Default Credentials to authorize and authenticate the client.
	CredentialsFile string
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
	AWSConfig
}

// Datastore configures access to a Google Cloud Datastore database backend.
type Datastore struct {
	Project string
	// If no account file is provided then Funnel will try to use Google Application
	// Default Credentials to authorize and authenticate the client.
	CredentialsFile string
}

// DynamoDB describes the configuration for Amazon DynamoDB backed processes
// such as the event writer and server.
type DynamoDB struct {
	TableBasename string
	AWSConfig
}

// LocalStorage describes the directories Funnel can read from and write to
type LocalStorage struct {
	Disabled    bool
	AllowedDirs []string
}

// Valid validates the LocalStorage configuration
func (l LocalStorage) Valid() bool {
	return !l.Disabled && len(l.AllowedDirs) > 0
}

// GSStorage describes configuration for the Google Cloud storage backend.
type GSStorage struct {
	Disabled bool
	// If no account file is provided then Funnel will try to use Google Application
	// Default Credentials to authorize and authenticate the client.
	CredentialsFile string
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
	// The maximum number of times to retry on error.
	// Defaults to 3.
	MaxRetries int
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

// HTTPStorage configures the http storage backend.
type HTTPStorage struct {
	Disabled bool
	// Timeout duration for http GET calls
	Timeout time.Duration
}

// Valid validates the HTTPStorage configuration.
func (h HTTPStorage) Valid() bool {
	return !h.Disabled
}
