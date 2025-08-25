package main

import (
	"os"

	"github.com/ohsu-comp-bio/funnel/cmd"
	"github.com/ohsu-comp-bio/funnel/logger"
)

// https://github.com/ohsu-comp-bio/funnel
func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		logger.PrintSimpleError(err)
		os.Exit(1)
	}
}
