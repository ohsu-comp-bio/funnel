package cmd

import (
	"funnel/config"
	"github.com/spf13/cobra"
)

var configFile string
var baseConf config.Config

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use: "funnel",
}

func init() {
	RootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "", "Config File")
	RootCmd.PersistentFlags().StringVar(&baseConf.HostName, "hostname", baseConf.HostName, "Host name or IP")
	RootCmd.PersistentFlags().StringVar(&baseConf.RPCPort, "rpc-port", baseConf.RPCPort, "RPC Port")
	RootCmd.PersistentFlags().StringVar(&baseConf.WorkDir, "work-dir", baseConf.WorkDir, "Working Directory")
	RootCmd.PersistentFlags().StringVar(&baseConf.LogLevel, "log-level", baseConf.LogLevel, "Level of logging")
	RootCmd.PersistentFlags().StringVar(&baseConf.LogPath, "log-path", baseConf.LogLevel, "File path to write logs to")
}
