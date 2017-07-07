package aws

import (
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var log = logger.New("aws cmd")

// Capture AWS Batch config (compute env, job queue, etc.)
var flagConf = DefaultConfig()

// Funnel's AWS Batch proxy passes the task message to this
// command as a JSON string via a CLI flag.
var rawTask string
var rawTaskFile string

func init() {
	f := runTaskCmd.Flags()
	f.StringVar(&rawTask, "task", "", "Task JSON")
	f.StringVar(&rawTaskFile, "task-file", "", "Task JSON file path")

	Cmd.AddCommand(deployCmd)
	d := deployCmd.Flags()
	// Capture AWS Batch config (compute env, job queue, etc.)
	d.StringSliceVar(
		&flagConf.ComputeEnv.Subnets,
		"subnet",
		flagConf.ComputeEnv.Subnets,
		"AWS Batch compute environment subnet ID")

	d.StringSliceVar(
		&flagConf.ComputeEnv.SecurityGroupIds,
		"security-group",
		flagConf.ComputeEnv.SecurityGroupIds,
		"AWS Batch compute environment security group ID")

	d.StringVar(&flagConf.Container, "container", flagConf.Container,
		"Funnel worker Docker container to run.")

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
		return deploy(flagConf)
	},
}

var runTaskCmd = &cobra.Command{
	Use: "runtask",
	RunE: func(cmd *cobra.Command, args []string) error {

		if rawTaskFile != "" {
			b, err := ioutil.ReadFile(rawTaskFile)
			if err != nil {
				return err
			}
			rawTask = string(b)
		}

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
