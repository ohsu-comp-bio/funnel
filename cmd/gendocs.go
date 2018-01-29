package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var genMarkdownCmd = &cobra.Command{
	Use:    "genmarkdown",
	Short:  "generate markdown formatted documentation for the funnel commands",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doc.GenMarkdownTree(RootCmd, "./funnel-cmd-docs")
	},
}

var genBashCompletionCmd = &cobra.Command{
	Use:    "genbash",
	Short:  "generate bash completions for the funnel commands",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletion(os.Stdout)
	},
}
