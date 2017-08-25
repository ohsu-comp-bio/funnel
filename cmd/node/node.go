package node

import (
	"github.com/spf13/cobra"
)

// Cmd represents the node command
var Cmd = &cobra.Command{
	Use:     "node",
	Aliases: []string{"nodes"},
	Short:   "Funnel node subcommands.",
}

func init() {
	Cmd.AddCommand(startCmd)
}
