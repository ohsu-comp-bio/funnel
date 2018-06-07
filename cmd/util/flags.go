package util

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/spf13/pflag"
)

func ConfigFlags(conf *config.Config) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)

	f.StringSliceVar(&conf.EventWriters, "EventWriters", conf.EventWriters, "component selectors")

	f.StringVar(&conf.Database, "Database", conf.Database, "")

	f.StringVar(&conf.Compute, "Compute", conf.Compute, "")

	f.StringVar(&conf.Server.ServiceName, "Server.ServiceName", conf.Server.ServiceName, "")

	f.StringVar(&conf.Server.HostName, "Server.HostName", conf.Server.HostName, "")

	f.StringVar(&conf.Server.HTTPPort, "Server.HTTPPort", conf.Server.HTTPPort, "")

	f.StringVar(&conf.Server.RPCPort, "Server.RPCPort", conf.Server.RPCPort, "")

	f.BoolVar(&conf.Server.DisableHTTPCache, "Server.DisableHTTPCache", conf.Server.DisableHTTPCache, "")

	f.StringVar(&conf.RPCClient.BasicCredential.User, "RPCClient.BasicCredential.User", conf.RPCClient.BasicCredential.User, "")

	f.StringVar(&conf.RPCClient.BasicCredential.Password, "RPCClient.BasicCredential.Password", conf.RPCClient.BasicCredential.Password, "")

	f.StringVar(&conf.RPCClient.ServerAddress, "RPCClient.ServerAddress", conf.RPCClient.ServerAddress, "")

	f.Var(&conf.RPCClient.Timeout, "RPCClient.Timeout", "The timeout to use for making RPC client connections in nanoseconds This timeout is Only enforced when used in conjunction with the grpc.WithBlock dial option.")

	f.UintVar(&conf.RPCClient.MaxRetries, "RPCClient.MaxRetries", conf.RPCClient.MaxRetries, "The maximum number of times that a request will be retried for failures.")

	f.Var(&conf.Scheduler.ScheduleRate, "Scheduler.ScheduleRate", "How often to run a scheduler iteration.")

	f.IntVar(&conf.Scheduler.ScheduleChunk, "Scheduler.ScheduleChunk", conf.Scheduler.ScheduleChunk, "How many tasks to schedule in one iteration.")

	f.Var(&conf.Scheduler.NodePingTimeout, "Scheduler.NodePingTimeout", "How long to wait for a node ping before marking it as dead")

	f.Var(&conf.Scheduler.NodeInitTimeout, "Scheduler.NodeInitTimeout", "How long to wait for node initialization before marking it dead")

	f.Var(&conf.Scheduler.NodeDeadTimeout, "Scheduler.NodeDeadTimeout", "How long to wait before deleting a dead node from the DB.")

	f.StringVar(&conf.Node.ID, "Node.ID", conf.Node.ID, "")

	f.Uint32Var(&conf.Node.Resources.Cpus, "Node.Resources.Cpus", conf.Node.Resources.Cpus, "A Node will automatically try to detect what resources are available to it.")

	f.Float64Var(&conf.Node.Resources.RamGb, "Node.Resources.RamGb", conf.Node.Resources.RamGb, "A Node will automatically try to detect what resources are available to it.")

	f.Float64Var(&conf.Node.Resources.DiskGb, "Node.Resources.DiskGb", conf.Node.Resources.DiskGb, "A Node will automatically try to detect what resources are available to it.")

	f.Var(&conf.Node.Timeout, "Node.Timeout", "If the node has been idle for longer than the timeout, it will shut down.")

	f.Var(&conf.Node.UpdateRate, "Node.UpdateRate", "How often the node sends update requests to the server.")

	f.StringVar(&conf.Worker.WorkDir, "Worker.WorkDir", conf.Worker.WorkDir, "Directory to write task files to")

	f.Var(&conf.Worker.PollingRate, "Worker.PollingRate", "How often the worker should poll for cancel signals")

	f.Var(&conf.Worker.LogUpdateRate, "Worker.LogUpdateRate", "How often to update stdout/stderr log fields.")

	f.Int64Var(&conf.Worker.LogTailSize, "Worker.LogTailSize", conf.Worker.LogTailSize, "Max bytes of stdout/stderr to store in the database.")

	f.BoolVar(&conf.Worker.LeaveWorkDir, "Worker.LeaveWorkDir", conf.Worker.LeaveWorkDir, "Normally the worker cleans up its working directory after executing.")

	f.StringVar(&conf.Logger.Level, "Logger.Level", conf.Logger.Level, "")

	f.StringVar(&conf.Logger.Formatter, "Logger.Formatter", conf.Logger.Formatter, "")

	f.StringVar(&conf.Logger.OutputFile, "Logger.OutputFile", conf.Logger.OutputFile, "")

	f.BoolVar(&conf.Logger.JSONFormat.DisableTimestamp, "Logger.JSONFormat.DisableTimestamp", conf.Logger.JSONFormat.DisableTimestamp, "")

	f.StringVar(&conf.Logger.JSONFormat.TimestampFormat, "Logger.JSONFormat.TimestampFormat", conf.Logger.JSONFormat.TimestampFormat, "")

	f.BoolVar(&conf.Logger.TextFormat.ForceColors, "Logger.TextFormat.ForceColors", conf.Logger.TextFormat.ForceColors, "Set to true to bypass checking for a TTY before outputting colors.")

	f.BoolVar(&conf.Logger.TextFormat.DisableColors, "Logger.TextFormat.DisableColors", conf.Logger.TextFormat.DisableColors, "Force disabling colors.")

	f.BoolVar(&conf.Logger.TextFormat.DisableTimestamp, "Logger.TextFormat.DisableTimestamp", conf.Logger.TextFormat.DisableTimestamp, "Disable timestamp logging.")

	f.BoolVar(&conf.Logger.TextFormat.FullTimestamp, "Logger.TextFormat.FullTimestamp", conf.Logger.TextFormat.FullTimestamp, "Enable logging the full timestamp when a TTY is attached instead of just the time passed since beginning of execution.")

	f.StringVar(&conf.Logger.TextFormat.TimestampFormat, "Logger.TextFormat.TimestampFormat", conf.Logger.TextFormat.TimestampFormat, "TimestampFormat to use for display when a full timestamp is printed")

	f.BoolVar(&conf.Logger.TextFormat.DisableSorting, "Logger.TextFormat.DisableSorting", conf.Logger.TextFormat.DisableSorting, "The fields are sorted by default for a consistent output.")

	f.StringVar(&conf.Logger.TextFormat.Indent, "Logger.TextFormat.Indent", conf.Logger.TextFormat.Indent, "")

	f.StringVar(&conf.BoltDB.Path, "BoltDB.Path", conf.BoltDB.Path, "")

	f.StringVar(&conf.Badger.Path, "Badger.Path", conf.Badger.Path, "Path to database directory.")

	f.StringVar(&conf.DynamoDB.TableBasename, "DynamoDB.TableBasename", conf.DynamoDB.TableBasename, "")

	f.StringVar(&conf.DynamoDB.AWSConfig.Endpoint, "DynamoDB.AWSConfig.Endpoint", conf.DynamoDB.AWSConfig.Endpoint, "An optional endpoint URL (hostname only or fully qualified URI) that overrides the default generated endpoint for a client.")

	f.StringVar(&conf.DynamoDB.AWSConfig.Region, "DynamoDB.AWSConfig.Region", conf.DynamoDB.AWSConfig.Region, "The region to send requests to.")

	f.IntVar(&conf.DynamoDB.AWSConfig.MaxRetries, "DynamoDB.AWSConfig.MaxRetries", conf.DynamoDB.AWSConfig.MaxRetries, "The maximum number of times that a request will be retried for failures.")

	f.StringVar(&conf.DynamoDB.AWSConfig.Key, "DynamoDB.AWSConfig.Key", conf.DynamoDB.AWSConfig.Key, "If both the key and secret are empty AWS credentials will be read from the environment.")

	f.StringVar(&conf.DynamoDB.AWSConfig.Secret, "DynamoDB.AWSConfig.Secret", conf.DynamoDB.AWSConfig.Secret, "")

	f.StringVar(&conf.Elastic.IndexPrefix, "Elastic.IndexPrefix", conf.Elastic.IndexPrefix, "")

	f.StringVar(&conf.Elastic.URL, "Elastic.URL", conf.Elastic.URL, "")

	f.StringSliceVar(&conf.MongoDB.Addrs, "MongoDB.Addrs", conf.MongoDB.Addrs, "Addrs holds the addresses for the seed servers.")

	f.StringVar(&conf.MongoDB.Database, "MongoDB.Database", conf.MongoDB.Database, "Database is the database name used within MongoDB to store funnel data.")

	f.Var(&conf.MongoDB.Timeout, "MongoDB.Timeout", "Timeout is the amount of time to wait for a server to respond when first connecting and on follow up operations in the session.")

	f.StringVar(&conf.MongoDB.Username, "MongoDB.Username", conf.MongoDB.Username, "Username and Password inform the credentials for the initial authentication done on the database defined by the Database field.")

	f.StringVar(&conf.MongoDB.Password, "MongoDB.Password", conf.MongoDB.Password, "")

	f.StringSliceVar(&conf.Kafka.Servers, "Kafka.Servers", conf.Kafka.Servers, "")

	f.StringVar(&conf.Kafka.Topic, "Kafka.Topic", conf.Kafka.Topic, "")

	f.StringVar(&conf.PubSub.Topic, "PubSub.Topic", conf.PubSub.Topic, "")

	f.StringVar(&conf.PubSub.Project, "PubSub.Project", conf.PubSub.Project, "")

	f.StringVar(&conf.PubSub.CredentialsFile, "PubSub.CredentialsFile", conf.PubSub.CredentialsFile, "If no account file is provided then Funnel will try to use Google Application Default Credentials to authorize and authenticate the client.")

	f.StringVar(&conf.Datastore.Project, "Datastore.Project", conf.Datastore.Project, "")

	f.StringVar(&conf.Datastore.CredentialsFile, "Datastore.CredentialsFile", conf.Datastore.CredentialsFile, "If no account file is provided then Funnel will try to use Google Application Default Credentials to authorize and authenticate the client.")

	f.BoolVar(&conf.HTCondor.DisableReconciler, "HTCondor.DisableReconciler", conf.HTCondor.DisableReconciler, "Turn off task state reconciler.")

	f.Var(&conf.HTCondor.ReconcileRate, "HTCondor.ReconcileRate", "ReconcileRate is how often the compute backend compares states in Funnel's backend to those reported by the backend")

	f.StringVar(&conf.HTCondor.Template, "HTCondor.Template", conf.HTCondor.Template, "")

	f.BoolVar(&conf.Slurm.DisableReconciler, "Slurm.DisableReconciler", conf.Slurm.DisableReconciler, "Turn off task state reconciler.")

	f.Var(&conf.Slurm.ReconcileRate, "Slurm.ReconcileRate", "ReconcileRate is how often the compute backend compares states in Funnel's backend to those reported by the backend")

	f.StringVar(&conf.Slurm.Template, "Slurm.Template", conf.Slurm.Template, "")

	f.BoolVar(&conf.PBS.DisableReconciler, "PBS.DisableReconciler", conf.PBS.DisableReconciler, "Turn off task state reconciler.")

	f.Var(&conf.PBS.ReconcileRate, "PBS.ReconcileRate", "ReconcileRate is how often the compute backend compares states in Funnel's backend to those reported by the backend")

	f.StringVar(&conf.PBS.Template, "PBS.Template", conf.PBS.Template, "")

	f.StringVar(&conf.GridEngine.Template, "GridEngine.Template", conf.GridEngine.Template, "")

	f.StringVar(&conf.AWSBatch.JobDefinition, "AWSBatch.JobDefinition", conf.AWSBatch.JobDefinition, "JobDefinition can be either a name or the Amazon Resource Name (ARN).")

	f.StringVar(&conf.AWSBatch.JobQueue, "AWSBatch.JobQueue", conf.AWSBatch.JobQueue, "JobQueue can be either a name or the Amazon Resource Name (ARN).")

	f.BoolVar(&conf.AWSBatch.DisableReconciler, "AWSBatch.DisableReconciler", conf.AWSBatch.DisableReconciler, "Turn off task state reconciler.")

	f.Var(&conf.AWSBatch.ReconcileRate, "AWSBatch.ReconcileRate", "ReconcileRate is how often the compute backend compares states in Funnel's backend to those reported by AWS Batch")

	f.StringVar(&conf.AWSBatch.AWSConfig.Endpoint, "AWSBatch.AWSConfig.Endpoint", conf.AWSBatch.AWSConfig.Endpoint, "An optional endpoint URL (hostname only or fully qualified URI) that overrides the default generated endpoint for a client.")

	f.StringVar(&conf.AWSBatch.AWSConfig.Region, "AWSBatch.AWSConfig.Region", conf.AWSBatch.AWSConfig.Region, "The region to send requests to.")

	f.IntVar(&conf.AWSBatch.AWSConfig.MaxRetries, "AWSBatch.AWSConfig.MaxRetries", conf.AWSBatch.AWSConfig.MaxRetries, "The maximum number of times that a request will be retried for failures.")

	f.StringVar(&conf.AWSBatch.AWSConfig.Key, "AWSBatch.AWSConfig.Key", conf.AWSBatch.AWSConfig.Key, "If both the key and secret are empty AWS credentials will be read from the environment.")

	f.StringVar(&conf.AWSBatch.AWSConfig.Secret, "AWSBatch.AWSConfig.Secret", conf.AWSBatch.AWSConfig.Secret, "")

	f.BoolVar(&conf.LocalStorage.Disabled, "LocalStorage.Disabled", conf.LocalStorage.Disabled, "")

	f.StringSliceVar(&conf.LocalStorage.AllowedDirs, "LocalStorage.AllowedDirs", conf.LocalStorage.AllowedDirs, "")

	f.BoolVar(&conf.AmazonS3.Disabled, "AmazonS3.Disabled", conf.AmazonS3.Disabled, "")

	f.StringVar(&conf.AmazonS3.AWSConfig.Endpoint, "AmazonS3.AWSConfig.Endpoint", conf.AmazonS3.AWSConfig.Endpoint, "An optional endpoint URL (hostname only or fully qualified URI) that overrides the default generated endpoint for a client.")

	f.StringVar(&conf.AmazonS3.AWSConfig.Region, "AmazonS3.AWSConfig.Region", conf.AmazonS3.AWSConfig.Region, "The region to send requests to.")

	f.IntVar(&conf.AmazonS3.AWSConfig.MaxRetries, "AmazonS3.AWSConfig.MaxRetries", conf.AmazonS3.AWSConfig.MaxRetries, "The maximum number of times that a request will be retried for failures.")

	f.StringVar(&conf.AmazonS3.AWSConfig.Key, "AmazonS3.AWSConfig.Key", conf.AmazonS3.AWSConfig.Key, "If both the key and secret are empty AWS credentials will be read from the environment.")

	f.StringVar(&conf.AmazonS3.AWSConfig.Secret, "AmazonS3.AWSConfig.Secret", conf.AmazonS3.AWSConfig.Secret, "")

	f.BoolVar(&conf.GoogleStorage.Disabled, "GoogleStorage.Disabled", conf.GoogleStorage.Disabled, "")

	f.StringVar(&conf.GoogleStorage.CredentialsFile, "GoogleStorage.CredentialsFile", conf.GoogleStorage.CredentialsFile, "If no account file is provided then Funnel will try to use Google Application Default Credentials to authorize and authenticate the client.")

	f.BoolVar(&conf.Swift.Disabled, "Swift.Disabled", conf.Swift.Disabled, "")

	f.StringVar(&conf.Swift.UserName, "Swift.UserName", conf.Swift.UserName, "")

	f.StringVar(&conf.Swift.Password, "Swift.Password", conf.Swift.Password, "")

	f.StringVar(&conf.Swift.AuthURL, "Swift.AuthURL", conf.Swift.AuthURL, "")

	f.StringVar(&conf.Swift.TenantName, "Swift.TenantName", conf.Swift.TenantName, "")

	f.StringVar(&conf.Swift.TenantID, "Swift.TenantID", conf.Swift.TenantID, "")

	f.StringVar(&conf.Swift.RegionName, "Swift.RegionName", conf.Swift.RegionName, "")

	f.Int64Var(&conf.Swift.ChunkSizeBytes, "Swift.ChunkSizeBytes", conf.Swift.ChunkSizeBytes, "Size of chunks to use for large object creation.")

	f.IntVar(&conf.Swift.MaxRetries, "Swift.MaxRetries", conf.Swift.MaxRetries, "The maximum number of times to retry on error.")

	f.BoolVar(&conf.HTTPStorage.Disabled, "HTTPStorage.Disabled", conf.HTTPStorage.Disabled, "")

	f.Var(&conf.HTTPStorage.Timeout, "HTTPStorage.Timeout", "Timeout duration for http GET calls")

	return f
}
