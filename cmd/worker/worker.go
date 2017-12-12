package worker

import (
	"context"
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
	Run func(ctx context.Context, conf config.Config, taskID string, log *logger.Logger) error
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
	f.AddFlagSet(util.WorkerFlags(&flagConf, &configFile, &serverAddress))

	run := &cobra.Command{
		Use:   "run",
		Short: "Run a task directly, bypassing the server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if taskID == "" {
				return fmt.Errorf("no taskID was provided")
			}
			log := logger.NewLogger("worker", conf.Logger)
			return hooks.Run(context.Background(), conf, taskID, log)
		},
	}
	f = run.Flags()
	f.StringVar(&taskID, "task-id", "", "Task ID")

	cmd.AddCommand(run)

	return cmd, hooks
}
