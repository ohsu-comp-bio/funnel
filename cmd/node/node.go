package node

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/util"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/spf13/cobra"
)

// NewCommand returns the node command
func NewCommand() *cobra.Command {
	cmd, _ := newCommandHooks()
	return cmd
}

type hooks struct {
	Run func(conf config.Config) error
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

	cmd := &cobra.Command{
		Use:     "node",
		Aliases: []string{"nodes"},
		Short:   "Funnel node subcommands.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			conf, err = util.MergeConfigFileWithFlags(configFile, flagConf)
			if err != nil {
				return fmt.Errorf("error processing config: %v", err)
			}

			return nil
		},
	}

	nodeFlags := util.ServerFlags(&flagConf, &configFile)
	cmd.SetGlobalNormalizationFunc(util.NormalizeFlags)
	f := cmd.PersistentFlags()
	f.AddFlagSet(nodeFlags)

	run := &cobra.Command{
		Use:   "run",
		Short: "Run a Funnel node.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return hooks.Run(conf)
		},
	}

	cmd.AddCommand(run)

	return cmd, hooks
}
