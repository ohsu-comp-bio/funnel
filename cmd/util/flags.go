package util

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/spf13/pflag"
)

// ServerFlags returns a new flag set for configuring a Funnel server
func ServerFlags(flagConf *config.Config, configFile *string) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	// Disable sorting in order to visit flags in primordial order below.
	f.SortFlags = false

	f.StringVarP(configFile, "config", "c", *configFile, "Config File")

	f.AddFlagSet(selectorFlags(flagConf))
	f.AddFlagSet(serverFlags(flagConf))
	f.AddFlagSet(workerFlags(flagConf))
	f.AddFlagSet(nodeFlags(flagConf))
	f.AddFlagSet(loggerFlags(flagConf))
	f.AddFlagSet(dbFlags(flagConf))

	return f
}

// WorkerFlags returns a new flag set for configuring a Funnel worker
func WorkerFlags(flagConf *config.Config, configFile *string, serverAddress *string) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	// Disable sorting in order to visit flags in primordial order below.
	f.SortFlags = false

	f.StringVarP(configFile, "config", "c", *configFile, "Config File")
	f.StringVarP(serverAddress, "server", "s", *serverAddress, "RPC address of Funnel server")

	f.AddFlagSet(selectorFlags(flagConf))
	f.AddFlagSet(workerFlags(flagConf))
	f.AddFlagSet(loggerFlags(flagConf))
	f.AddFlagSet(dbFlags(flagConf))

	return f
}

// NodeFlags returns a new flag set for configuring a Funnel node
func NodeFlags(flagConf *config.Config, configFile *string, serverAddress *string) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	// Disable sorting in order to visit flags in primordial order below.
	f.SortFlags = false

	f.StringVarP(configFile, "config", "c", *configFile, "Config File")
	f.StringVarP(serverAddress, "server", "s", *serverAddress, "RPC address of Funnel server")

	f.AddFlagSet(nodeFlags(flagConf))
	f.AddFlagSet(loggerFlags(flagConf))

	return f
}

func selectorFlags(flagConf *config.Config) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	// Disable sorting in order to visit flags in primordial order below.
	f.SortFlags = false

	f.StringVar(&flagConf.Database, "Database", flagConf.Database, "Name of database backed to use.")
	f.StringVar(&flagConf.Compute, "Compute", flagConf.Compute, "Name of compute backed to use.")
	f.StringSliceVar(&flagConf.EventWriters, "EventWriter", flagConf.EventWriters, "Name of an event writer backend to use. This flag can be used multiple times")

	return f
}

func serverFlags(flagConf *config.Config) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	// Disable sorting in order to visit flags in primordial order below.
	f.SortFlags = false

	f.StringVar(&flagConf.Server.ServiceName, "Server.ServiceName", flagConf.Server.ServiceName, "Host name or IP")
	f.StringVar(&flagConf.Server.HostName, "Server.Hostname", flagConf.Server.HostName, "Host name or IP")
	f.StringVar(&flagConf.Server.RPCPort, "Server.RPCPort", flagConf.Server.RPCPort, "RPC Port")
	f.StringVar(&flagConf.Server.HTTPPort, "Server.HTTPPort", flagConf.Server.HTTPPort, "HTTP Port")

	return f
}

func workerFlags(flagConf *config.Config) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	// Disable sorting in order to visit flags in primordial order below.
	f.SortFlags = false

	f.StringVar(&flagConf.Worker.WorkDir, "Worker.WorkDir", flagConf.Worker.WorkDir, "Working Directory")
	f.DurationVar(&flagConf.Worker.UpdateRate, "Worker.UpdateRate", flagConf.Worker.UpdateRate, "Task log update rate")
	f.Int64Var(&flagConf.Worker.BufferSize, "Worker.BufferSize", flagConf.Worker.BufferSize, "Max bytes to store for stderr/stdout")
	f.StringVar(&flagConf.Worker.TaskReader, "Worker.TaskReader", flagConf.Worker.TaskReader, "Name of the task reader backend to use")

	return f
}

func nodeFlags(flagConf *config.Config) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	// Disable sorting in order to visit flags in primordial order below.
	f.SortFlags = false

	f.StringVar(&flagConf.Node.ID, "Node.ID", flagConf.Node.ID, "Node ID")
	f.DurationVar(&flagConf.Node.Timeout, "Node.Timeout", flagConf.Node.Timeout, "Node timeout in seconds")
	f.DurationVar(&flagConf.Node.UpdateRate, "Node.UpdateRate", flagConf.Node.UpdateRate, "Node update rate")
	f.DurationVar(&flagConf.Node.UpdateTimeout, "Node.UpdateTimeout", flagConf.Node.UpdateTimeout, "Node update timeout")
	f.Uint32Var(&flagConf.Node.Resources.Cpus, "Node.Resources.Cpus", flagConf.Node.Resources.Cpus, "Cpus available to Node")
	f.Float64Var(&flagConf.Node.Resources.RamGb, "Node.Resources.RamGb", flagConf.Node.Resources.RamGb, "Ram (GB) available to Node")
	f.Float64Var(&flagConf.Node.Resources.DiskGb, "Node.Resources.DiskGb", flagConf.Node.Resources.DiskGb, "Free disk (GB) available to Node")

	return f
}

func loggerFlags(flagConf *config.Config) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	// Disable sorting in order to visit flags in primordial order below.
	f.SortFlags = false

	f.StringVar(&flagConf.Logger.Level, "Logger.Level", flagConf.Logger.Level, "Level of logging")
	f.StringVar(&flagConf.Logger.OutputFile, "Logger.OutputFile", flagConf.Logger.OutputFile, "File path to write logs to")

	return f
}

func dbFlags(flagConf *config.Config) *pflag.FlagSet {
	f := pflag.NewFlagSet("", pflag.ContinueOnError)
	// Disable sorting in order to visit flags in primordial order below.
	f.SortFlags = false

	// boltdb
	f.StringVar(&flagConf.BoltDB.Path, "BoltDB.Path", flagConf.BoltDB.Path, "Path to BoltDB database")

	// dynamodb
	f.StringVar(&flagConf.DynamoDB.Region, "DynamoDB.Region", flagConf.DynamoDB.Region, "AWS region of DynamoDB tables")
	f.StringVar(&flagConf.DynamoDB.TableBasename, "DynamoDB.TableBasename", flagConf.DynamoDB.TableBasename, "Basename of DynamoDB tables")

	// elastic
	f.StringVar(&flagConf.Elastic.IndexPrefix, "Elastic.IndexPrefix", flagConf.Elastic.IndexPrefix, "Prefix to use for Elasticsearch indices")
	f.StringVar(&flagConf.Elastic.URL, "Elastic.URL", flagConf.Elastic.URL, "Elasticsearch URL")

	// mongodb
	f.StringSliceVar(&flagConf.MongoDB.Addrs, "MongoDB.Addr", flagConf.MongoDB.Addrs, "Address of a MongoDB seed server. This flag can be used multiple times")
	f.StringVar(&flagConf.MongoDB.Database, "MongoDB.Database", flagConf.MongoDB.Database, "Database name in MongoDB")

	// kafka
	f.StringSliceVar(&flagConf.Kafka.Servers, "Kafka.Server", flagConf.Kafka.Servers, "Address of a Kafka server. This flag can be used multiple times")
	f.StringVar(&flagConf.Kafka.Topic, "Kafka.Topic", flagConf.Kafka.Topic, "Kafka topic to write events to")

	return f
}
