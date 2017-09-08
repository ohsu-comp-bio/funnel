package version

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/spf13/cobra"
)

//go:generate /bin/sh generate_data.sh

// Cmd represents the "version" command
var Cmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("git commit:", GitCommit)
		fmt.Println("git branch:", GitBranch)
		fmt.Println("build date:", BuildDate)
		fmt.Println("version:", Version)
	},
}

// Log logs build and version information to the given logger.
func Log(l logger.Logger) {
	l.Info("Version", "GitCommit", GitCommit, "GitBranch", GitBranch, "BuildDate", BuildDate,
		"Version", Version)
}
