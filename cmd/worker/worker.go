package worker

import (
	"context"
	"fmt"
	"syscall"
	"time"

	cmdutil "github.com/ohsu-comp-bio/funnel/cmd/util"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/spf13/cobra"
)

// NewCommand returns the worker command
func NewCommand() *cobra.Command {
	cmd, _ := newCommandHooks()
	return cmd
}

type hooks struct {
	Run func(ctx context.Context, conf config.Config, log *logger.Logger, taskID string) error
}

func newCommandHooks() (*cobra.Command, *hooks) {
	hooks := &hooks{
		Run: Run,
	}

	var (
		configFile string
		conf       config.Config
		flagConf   config.Config
		taskID     string
	)
	workerFlags := cmdutil.ConfigFlags(&flagConf)

	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Funnel worker commands.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			cmdutil.LookupEnv(workerFlags)
			conf, err = cmdutil.MergeConfigFileWithFlags(configFile, flagConf)
			if err != nil {
				return fmt.Errorf("error processing config: %v", err)
			}

			return nil
		},
	}
	cmd.SetGlobalNormalizationFunc(cmdutil.NormalizeFlags)
	f := cmd.PersistentFlags()
	f.StringVarP(&configFile, "config", "c", configFile, "Config File")
	f.AddFlagSet(workerFlags)

	run := &cobra.Command{
		Use:   "run",
		Short: "Run a task directly, bypassing the server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if taskID == "" {
				return fmt.Errorf("no taskID was provided")
			}

			log := logger.NewLogger("worker", conf.Logger)
			logger.SetGRPCLogger(log)
			ctx, cancel := context.WithCancel(context.Background())
			ctx = util.SignalContext(ctx, time.Millisecond*500, syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			return hooks.Run(ctx, conf, log, taskID)
		},
	}

	f = run.Flags()
	f.StringVarP(&taskID, "taskID", "t", taskID, "Task ID")
	cmd.AddCommand(run)

	return cmd, hooks
}
