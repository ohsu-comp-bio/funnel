package server

import (
	"context"
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
	Run func(ctx context.Context, conf config.Config) error
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
		Use:   "server",
		Short: "Funnel server commands.",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error

			conf, err = util.MergeConfigFileWithFlags(configFile, flagConf)
			if err != nil {
				return fmt.Errorf("error processing config: %v", err)
			}

			return nil
		},
	}
	f := cmd.PersistentFlags()
	f.StringVarP(&configFile, "config", "c", "", "Config File")
	f.StringVar(&flagConf.Server.HostName, "hostname", flagConf.Server.HostName, "Host name or IP")
	f.StringVar(&flagConf.Server.RPCPort, "rpc-port", flagConf.Server.RPCPort, "RPC Port")
	f.StringVar(&flagConf.Server.HTTPPort, "http-port", flagConf.Server.HTTPPort, "HTTP Port")
	f.StringVar(&flagConf.Server.Logger.Level, "log-level", flagConf.Server.Logger.Level, "Level of logging")
	f.StringVar(&flagConf.Server.Logger.OutputFile, "log-path", flagConf.Server.Logger.OutputFile, "File path to write logs to")
	f.StringVar(&flagConf.Server.Database, "database", flagConf.Server.Database, "Name of database backend to enable")
	f.StringVar(&flagConf.Backend, "backend", flagConf.Backend, "Name of compute backend to enable")

	run := &cobra.Command{
		Use:   "run",
		Short: "Runs a Funnel server.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return hooks.Run(context.Background(), conf)
		},
	}

	cmd.AddCommand(run)

	return cmd, hooks
}
