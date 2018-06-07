package node

import (
	"context"
	"fmt"

	cmdutil "github.com/ohsu-comp-bio/funnel/cmd/util"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
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

	nodeFlags := cmdutil.ConfigFlags(&flagConf)

	cmd := &cobra.Command{
		Use:     "node",
		Aliases: []string{"nodes"},
		Short:   "Funnel node subcommands.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			cmdutil.LookupEnv(nodeFlags)
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
	f.AddFlagSet(nodeFlags)

	run := &cobra.Command{
		Use:   "run",
		Short: "Run a Funnel node.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logger.NewLogger("node", conf.Logger)
			logger.SetGRPCLogger(log)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			return hooks.Run(ctx, conf, log)
		},
	}

	cmd.AddCommand(run)

	return cmd, hooks
}
