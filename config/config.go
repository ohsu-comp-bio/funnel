// Package config contains Funnel configuration structures and defaults.
package config

import (
	"os"
)

// Config describes configuration for Funnel.
type Config struct {
	// component selectors
	EventWriters []string
	Database     string
	Compute      string
	// funnel components
	Server    Server
	RPCClient RPCClient
	Scheduler Scheduler
	Node      Node
	Worker    Worker
	Logger    logger.Config
	// databases / event handlers
	BoltDB    BoltDB
	Badger    Badger
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
	GridEngine struct {
		Template     string
		TemplateFile string
	}
	AWSBatch   AWSBatch
	GCPBatch   GCPBatch
	Kubernetes Kubernetes
	// storage
	LocalStorage  LocalStorage
	AmazonS3      AmazonS3Storage
	GenericS3     []GenericS3Storage
	GoogleStorage GoogleCloudStorage
	Swift         SwiftStorage
	HTTPStorage   HTTPStorage
	FTPStorage    FTPStorage
}

// BasicCredential describes a username and password for use with Funnel's basic auth.
type BasicCredential struct {
	User     string
	Password string
	Admin    bool
}

type OidcAuth struct {
	ServiceConfigURL string
	ClientId         string
	ClientSecret     string
	RedirectURL      string
	RequireScope     string
	RequireAudience  string
	Admins           []string
}

// RPCClient describes configuration for gRPC clients
type RPCClient struct {
	BasicCredential
	ServerAddress string
	// The timeout to use for making RPC client connections in nanoseconds
	// This timeout is Only enforced when used in conjunction with the
	// grpc.WithBlock dial option.
	Timeout Duration
	// The maximum number of times that a request will be retried for failures.
	// Time between retries follows an exponential backoff starting at 5 seconds
	// up to 1 minute
	MaxRetries uint
}

// Server describes configuration for the server.
type Server struct {
	ServiceName      string
	HostName         string
	HTTPPort         string
	RPCPort          string
	BasicAuth        []BasicCredential
	OidcAuth         OidcAuth
	DisableHTTPCache bool

	// Defines task access and visibility by options:
	// "All" (default) – all tasks are visible to everyone
	// "Owner" - tasks are visible to the users who created them
	// "OwnerOrAdmin" - extends "Owner" by allowing Admin-users see everything
	TaskAccess string
}

// HTTPAddress returns the HTTP address based on HostName and HTTPPort
func (c *Server) HTTPAddress() string {
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
	ScheduleRate Duration
	// How many tasks to schedule in one iteration.
	ScheduleChunk int
	// How long to wait for a node ping before marking it as dead
	NodePingTimeout Duration
	// How long to wait for node initialization before marking it dead
	NodeInitTimeout Duration
	// How long to wait before deleting a dead node from the DB.
	NodeDeadTimeout Duration
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
	Timeout Duration
	// How often the node sends update requests to the server.
	UpdateRate Duration
	Metadata   map[string]string
}

// Worker contains worker configuration.
type Worker struct {
	// Directory to write task files to
	WorkDir string
	// Additional directory to symlink to the working directory.
	ScratchPath string
	// How often the worker should poll for cancel signals
	PollingRate Duration
	// How often to update stdout/stderr log fields.
	// Setting this to 0 will result in these fields being updated a single time
	// after the executor exits.
	LogUpdateRate Duration
	// Max bytes of stdout/stderr to store in the database.
	// Setting this to 0 turns off stdout/stderr logging.
	LogTailSize int64
	// Normally the worker cleans up its working directory after executing.
	// This option disables that behavior.
	LeaveWorkDir bool
	// Limit the number of concurrent downloads/uploads
	MaxParallelTransfers int
	// Container engine to use for executing tasks.
	Container ContainerConfig
	// Command to use for the container engine.
	// This can be used to override the default command used to run containers.
	DriverCommand string
}

type ContainerConfig struct {
	Id              string
	Image           string
	Name            string
	Command         []string
	Workdir         string
	RemoveContainer bool
	Env             map[string]string
	Stdin           io.Reader
	Stdout          io.Writer
	Stderr          io.Writer
	DriverCommand   string
	RunCommand      string // template string
	PullCommand     string // template string
	StopCommand     string // template string
	EnableTags      bool
	Tags            map[string]string
}

// HPCBackend describes the configuration for a HPC scheduler backend such as
// HTCondor or Slurm.
type HPCBackend struct {
	// Turn off task state reconciler. When enabled, Funnel communicates with the HPC
	// scheduler to find tasks that are stuck in a queued state or errored and
	// updates the task state accordingly.
	DisableReconciler bool
	// ReconcileRate is how often the compute backend compares states in Funnel's backend
	// to those reported by the backend
	ReconcileRate Duration
	// Template is the template to use for submissions to the HPC backend
	Template string
	// TemplateFile is the path to the template to use for submissions to the HPC backend
	TemplateFile string
}

// BoltDB describes the configuration for the BoltDB embedded database.
type BoltDB struct {
	Path string
}

// Badger describes configuration for the Badger embedded database.
type Badger struct {
	// Path to database directory.
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
	Timeout Duration
	// Username and Password inform the credentials for the initial authentication
	// done on the database defined by the Database field.
	Username string
	Password string
}

// Elastic configures access to an Elasticsearch database.
type Elastic struct {
	IndexPrefix  string
	URL          string
	Username     string
	Password     string
	CloudID      string
	APIKey       string
	ServiceToken string
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
	// Don't automatically read AWS credentials from the environment.
	DisableAutoCredentialLoad bool
}

// AWSBatch describes the configuration for the AWS Batch compute backend.
type AWSBatch struct {
	// JobDefinition can be either a name or the Amazon Resource Name (ARN).
	JobDefinition string
	// JobQueue can be either a name or the Amazon Resource Name (ARN).
	JobQueue string
	// Turn off task state reconciler. When enabled, Funnel communicates with AWS Batch
	// to find tasks that never started and updates task state accordingly.
	DisableReconciler bool
	// ReconcileRate is how often the compute backend compares states in Funnel's backend
	// to those reported by AWS Batch
	ReconcileRate Duration
	AWSConfig
}

// Initial pass at Google Batch configuration (relying on existing AWS config as much as possible).
type GCPBatch struct {
	JobDefinition     string
	JobQueue          string
	DisableReconciler bool
	ReconcileRate     Duration
	// Required fields for GCP Batch Jobs (as "Parent") → projects/{project}/locations/{location}
	// https://pkg.go.dev/cloud.google.com/go/batch/apiv1/batchpb#CreateJobRequest
	Project  string
	Location string
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
func (l *LocalStorage) Valid() bool {
	return !l.Disabled && len(l.AllowedDirs) > 0
}

// Valid validates the Storage configuration.
func (g *GoogleCloudStorage) Valid() bool {
	return !g.Disabled
}

// Valid validates the AmazonS3Storage configuration
func (s *AmazonS3Storage) Valid() bool {
	creds := s.AWSConfig == nil || (s.AWSConfig.Key != "" && s.AWSConfig.Secret != "") || (s.AWSConfig.Key == "" && s.AWSConfig.Secret == "")
	return !s.Disabled && creds
}

// Valid validates the S3Storage configuration
func (s *GenericS3Storage) Valid() bool {
	return !s.Disabled && s.Key != "" && s.Secret != "" && s.Endpoint != ""
}

// Valid validates the SwiftStorage configuration.
func (s *SwiftStorage) Valid() bool {
	user := s.UserName != "" || os.Getenv("OS_USERNAME") != ""
	password := s.Password != "" || os.Getenv("OS_PASSWORD") != ""
	authURL := s.AuthURL != "" || os.Getenv("OS_AUTH_URL") != ""
	tenantName := s.TenantName != "" || os.Getenv("OS_TENANT_NAME") != "" || os.Getenv("OS_PROJECT_NAME") != ""
	tenantID := s.TenantID != "" || os.Getenv("OS_TENANT_ID") != "" || os.Getenv("OS_PROJECT_ID") != ""
	region := s.RegionName != "" || os.Getenv("OS_REGION_NAME") != ""

	valid := user && password && authURL && tenantName && tenantID && region

	return !s.Disabled && valid
}

// Valid validates the HTTPStorage configuration.
func (h *HTTPStorage) Valid() bool {
	return !h.Disabled
}

// Valid validates the FTPStorage configuration.
func (h *FTPStorage) Valid() bool {
	return !h.Disabled
}
