package worker

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/util"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
)

// NewCommand returns the worker command
func NewCommand() *cobra.Command {
	cmd, _ := newCommandHooks()
	return cmd
}

type hooks struct {
	Run func(conf config.Worker, taskID string, log *logger.Logger) error
}

func newCommandHooks() (*cobra.Command, *hooks) {
	hooks := &hooks{
		Run: Run,
	}

	var (
		configFile            string
		conf                  config.Config
		flagConf              config.Config
		serverAddress         string
		dynamodbRegion        string
		dynamodbTableBasename string
		taskID                string
	)

	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Funnel worker commands.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			flagConf, err = util.ParseServerAddressFlag(serverAddress, flagConf)
			if err != nil {
				return fmt.Errorf("error parsing the server address: %v", err)
			}

			conf, err = util.MergeConfigFileWithFlags(configFile, flagConf)
			if err != nil {
				return fmt.Errorf("error processing config: %v", err)
			}

			if dynamodbRegion != "" {
				conf.Worker.EventWriters.DynamoDB.AWS.Region = dynamodbRegion
				conf.Worker.TaskReaders.DynamoDB.AWS.Region = dynamodbRegion
			}
			if dynamodbTableBasename != "" {
				conf.Worker.EventWriters.DynamoDB.TableBasename = dynamodbTableBasename
				conf.Worker.TaskReaders.DynamoDB.TableBasename = dynamodbTableBasename
			}

			return nil
		},
	}
	f := cmd.PersistentFlags()
	f.StringVarP(&configFile, "config", "c", "", "Config File")
	f.StringVar(&flagConf.Worker.WorkDir, "WorkDir", flagConf.Worker.WorkDir, "Working Directory")
	f.StringVar(&flagConf.Worker.Logger.Level, "Logger.Level", flagConf.Worker.Logger.Level, "Level of logging")
	f.StringVar(&flagConf.Worker.Logger.OutputFile, "Logger.OutputFile", flagConf.Worker.Logger.OutputFile, "File path to write logs to")
	f.StringVar(&flagConf.Worker.TaskReader, "TaskReader", flagConf.Worker.TaskReader, "Name of the task reader backend to use")
	f.StringSliceVar(&flagConf.Worker.ActiveEventWriters, "ActiveEventWriters", flagConf.Worker.ActiveEventWriters, "Name of an event writer backend to use. This flag can be used multiple times")
	f.StringVar(&serverAddress, "RPC.ServerAddress", "", "RPC address of Funnel server - used by TaskReader and EventWriter")
	f.StringVar(&dynamodbRegion, "DynamoDB.Region", "", "AWS region of DynamoDB tables - used by TaskReader and EventWriter")
	f.StringVar(&dynamodbTableBasename, "DynamoDB.TableBasename", "", "Basename of DynamoDB tables - used by TaskReader and EventWriter")

	run := &cobra.Command{
		Use:   "run",
		Short: "Run a task directly, bypassing the server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if taskID == "" {
				return fmt.Errorf("no taskID was provided")
			}
			log := logger.NewLogger("worker", conf.Worker.Logger)
			return hooks.Run(conf.Worker, taskID, log)
		},
	}
	f = run.Flags()
	f.StringVar(&taskID, "task-id", "", "Task ID")

	cmd.AddCommand(run)

	return cmd, hooks
}
