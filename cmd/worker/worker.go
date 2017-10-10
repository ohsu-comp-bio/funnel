package worker

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/util"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/spf13/cobra"
)

// NewCommand returns the worker command
func NewCommand() *cobra.Command {
	cmd, _ := newCommandHooks()
	return cmd
}

type hooks struct {
	Run func(conf config.Worker, taskID string) error
}

func newCommandHooks() (*cobra.Command, *hooks) {
	hooks := &hooks{
		Run: Run,
	}

	var (
		configFile    string
		conf          config.Config
		flagConf      config.Config
		serverAddress string
		taskID        string
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

			return nil
		},
	}
	f := cmd.PersistentFlags()
	f.StringVarP(&configFile, "config", "c", "", "Config File")
	f.StringVar(&serverAddress, "server-address", "", "RPC address of Funnel server")
	f.StringVar(&flagConf.Worker.WorkDir, "work-dir", flagConf.Worker.WorkDir, "Working Directory")
	f.StringVar(&flagConf.Worker.Logger.Level, "log-level", flagConf.Worker.Logger.Level, "Level of logging")
	f.StringVar(&flagConf.Worker.Logger.OutputFile, "log-path", flagConf.Worker.Logger.OutputFile, "File path to write logs to")

	run := &cobra.Command{
		Use:   "run",
		Short: "Run a task directly, bypassing the server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if taskID == "" {
				return fmt.Errorf("no taskID was provided")
			}

			return hooks.Run(conf.Worker, taskID)
		},
	}
	f = run.Flags()
	f.StringVar(&taskID, "task-id", "", "Task ID")

	cmd.AddCommand(run)

	return cmd, hooks
}
