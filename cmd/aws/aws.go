package aws

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/spf13/cobra"
)

var log = logger.New("aws cmd")

// Funnel's AWS Batch proxy passes the task message to this
// command as a JSON string via a CLI flag.
var rawTask string

func init() {
	f := runTaskCmd.Flags()
	f.StringVar(&rawTask, "task", "", "Task JSON")
	Cmd.AddCommand(deployCmd)
	Cmd.AddCommand(runTaskCmd)
	Cmd.AddCommand(proxyCmd)
}

// Cmd is the aws command
var Cmd = &cobra.Command{
	Use: "aws",
}

var deployCmd = &cobra.Command{
	Use: "deploy",
	RunE: func(cmd *cobra.Command, args []string) error {
		return deploy()
	},
}

var runTaskCmd = &cobra.Command{
	Use: "runtask",
	RunE: func(cmd *cobra.Command, args []string) error {

		// Load tes.Task from raw string (comes from CLI flag).
		var task tes.Task
		err := jsonpb.UnmarshalString(rawTask, &task)
		if err != nil {
			return err
		}

		conf := config.DefaultConfig()
		log.Configure(conf.Logger)
		return runTask(&task, conf)
	},
}

var proxyCmd = &cobra.Command{
	Use: "proxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		conf := config.DefaultConfig()
		conf.Logger.Level = "debug"
		log.Configure(conf.Logger)
		return runProxy(conf)
	},
}
