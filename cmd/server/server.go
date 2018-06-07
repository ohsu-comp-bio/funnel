package server

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

// NewCommand returns the node command
func NewCommand() *cobra.Command {
	cmd, _ := newCommandHooks()
	return cmd
}

type hooks struct {
	Run func(ctx context.Context, conf config.Config, log *logger.Logger) error
}

func newCommandHooks() (*cobra.Command, *hooks) {
	hooks := &hooks{
		Run: Run,
	}

	var (
		configFile string
		conf       config.Config
		flagConf   config.Config
	)

	serverFlags := cmdutil.ConfigFlags(&flagConf)

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Funnel server commands.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			cmdutil.LookupEnv(serverFlags)
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
	f.AddFlagSet(serverFlags)

	run := &cobra.Command{
		Use:   "run",
		Short: "Runs a Funnel server.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger("server", conf.Logger)
			logger.SetGRPCLogger(log)
			ctx, cancel := context.WithCancel(context.Background())
			ctx = util.SignalContext(ctx, time.Millisecond*500, syscall.SIGINT, syscall.SIGTERM)
			defer cancel()
			return hooks.Run(ctx, conf, log)
		},
	}

	cmd.AddCommand(run)

	return cmd, hooks
}
