package worker

import (
	"context"
	"fmt"
	cmdutil "github.com/ohsu-comp-bio/funnel/cmd/util"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/spf13/cobra"
	"syscall"
)

// NewCommand returns the worker command
func NewCommand() *cobra.Command {
	cmd, _ := newCommandHooks()
	return cmd
}

type hooks struct {
	Run func(ctx context.Context, conf config.Config, taskID string, writer events.Writer, log *logger.Logger) error
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

	cmd := &cobra.Command{
		Use:   "worker",
		Short: "Funnel worker commands.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			conf, err = cmdutil.MergeConfigFileWithFlags(configFile, flagConf)
			if err != nil {
				return fmt.Errorf("error processing config: %v", err)
			}

			return nil
		},
	}
	workerFlags := cmdutil.WorkerFlags(&flagConf, &configFile)
	cmd.SetGlobalNormalizationFunc(cmdutil.NormalizeFlags)
	f := cmd.PersistentFlags()
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

			ctx, cancel := context.WithCancel(context.Background())
			ctx = util.SignalContext(ctx, syscall.SIGINT, syscall.SIGTERM)
			defer cancel()

			ew, err := NewWorkerEventWriter(ctx, conf, log)
			if err != nil {
				return err
			}

			return hooks.Run(ctx, conf, taskID, ew, log)
		},
	}

	f = run.Flags()
	f.StringVarP(&taskID, "taskID", "t", taskID, "Task ID")
	cmd.AddCommand(run)

	return cmd, hooks
}
