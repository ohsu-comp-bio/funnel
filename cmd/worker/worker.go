package worker

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/spf13/cobra"
)

var configFile string
var flagConf = config.Config{}

// Cmd represents the worker command
var Cmd = &cobra.Command{
	Use:   "worker",
	Short: "Funnel worker commands.",
}

func init() {
	flags := Cmd.PersistentFlags()
	flags.StringVarP(&configFile, "config", "c", "", "Config File")
	flags.StringVar(&flagConf.Worker.ServerAddress, "server-address", flagConf.Worker.ServerAddress, "RPC address of Funnel server")
	flags.StringVar(&flagConf.Worker.WorkDir, "work-dir", flagConf.Worker.WorkDir, "Working Directory")
	flags.StringVar(&flagConf.Worker.Logger.Level, "log-level", flagConf.Worker.Logger.Level, "Level of logging")
	flags.StringVar(&flagConf.Worker.Logger.OutputFile, "log-path", flagConf.Worker.Logger.OutputFile, "File path to write logs to")
	flags.StringVar(&flagConf.Worker.Logger.Formatter, "log-format", flagConf.Worker.Logger.Formatter, "Log format. ['json', 'text']")

	Cmd.AddCommand(runCmd)
}
