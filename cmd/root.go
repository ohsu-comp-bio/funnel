package cmd

import (
	"github.com/ohsu-comp-bio/funnel/cmd/examples"
	"github.com/ohsu-comp-bio/funnel/cmd/gce"
	"github.com/ohsu-comp-bio/funnel/cmd/node"
	"github.com/ohsu-comp-bio/funnel/cmd/run"
	"github.com/ohsu-comp-bio/funnel/cmd/server"
	"github.com/ohsu-comp-bio/funnel/cmd/task"
	"github.com/ohsu-comp-bio/funnel/cmd/termdash"
	"github.com/ohsu-comp-bio/funnel/cmd/version"
	"github.com/ohsu-comp-bio/funnel/cmd/worker"
	"github.com/spf13/cobra"
)

// RootCmd represents the root command
var RootCmd = &cobra.Command{
	Use:           "funnel",
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	RootCmd.AddCommand(examples.Cmd)
	RootCmd.AddCommand(gce.Cmd)
	RootCmd.AddCommand(genBashCompletionCmd)
	RootCmd.AddCommand(genMarkdownCmd)
	RootCmd.AddCommand(node.Cmd)
	RootCmd.AddCommand(run.Cmd)
	RootCmd.AddCommand(server.Cmd)
	RootCmd.AddCommand(task.Cmd)
	RootCmd.AddCommand(termdash.Cmd)
	RootCmd.AddCommand(version.Cmd)
	RootCmd.AddCommand(worker.Cmd)
}
