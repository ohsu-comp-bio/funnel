package main

import (
	"os"

	"github.com/ohsu-comp-bio/funnel/cmd"
	"github.com/ohsu-comp-bio/funnel/logger"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		logger.PrintSimpleError(err)
		os.Exit(1)
	}
}
