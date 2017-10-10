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
		configFile    string
		conf          config.Config
		flagConf      config.Config
		serverAddress string
	)

	cmd := &cobra.Command{
		Use:     "node",
		Aliases: []string{"nodes"},
		Short:   "Funnel node subcommands.",
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
	f.StringVar(&flagConf.Scheduler.Node.ID, "id", flagConf.Scheduler.Node.ID, "Node ID")
	f.StringVar(&serverAddress, "server-address", "", "Address of scheduler gRPC endpoint")
	f.DurationVar(&flagConf.Scheduler.Node.Timeout, "timeout", flagConf.Scheduler.Node.Timeout, "Timeout in seconds")
	f.StringVar(&flagConf.Scheduler.Node.WorkDir, "work-dir", flagConf.Scheduler.Node.WorkDir, "Working Directory")
	f.StringVar(&flagConf.Scheduler.Node.Logger.Level, "log-level", flagConf.Scheduler.Node.Logger.Level, "Level of logging")
	f.StringVar(&flagConf.Scheduler.Node.Logger.OutputFile, "log-path", flagConf.Scheduler.Node.Logger.OutputFile, "File path to write logs to")

	run := &cobra.Command{
		Use:   "run",
		Short: "Run a Funnel node.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return hooks.Run(conf)
		},
	}

	cmd.AddCommand(run)

	return cmd, hooks
}
