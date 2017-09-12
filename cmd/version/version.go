package version

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/version"
	"github.com/spf13/cobra"
)

// Cmd represents the "version" command
var Cmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		if version.GitCommit != "" {
			fmt.Println("git commit:", version.GitCommit)
		}
		if version.GitBranch != "" {
			fmt.Println("git branch:", version.GitBranch)
		}
		if version.GitUpstream != "" {
			fmt.Println("git upstream:", version.GitUpstream)
		}
		if version.BuildDate != "" {
			fmt.Println("build date:", version.BuildDate)
		}
		fmt.Println("version:", version.Version)
	},
}

// Log logs build and version information to the given logger.
func Log(l logger.Logger) {
	l.Info("Version", "GitCommit", version.GitCommit, "GitBranch", version.GitBranch,
		"BuildDate", version.BuildDate, "Version", version.Version)
}
