package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
	"os"
)

var markdownCmd = &cobra.Command{
	Use:    "genmarkdown",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return doc.GenMarkdownTree(RootCmd, "./funnel-cmd-docs")
	},
}

var genBashCompletionCmd = &cobra.Command{
	Use: "bash",
	Run: func(cmd *cobra.Command, args []string) {
		RootCmd.GenBashCompletion(os.Stdout)
	},
}
