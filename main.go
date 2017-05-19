package main

import (
	"github.com/ohsu-comp-bio/funnel/cmd"
	"github.com/ohsu-comp-bio/funnel/logger"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		logger.PrintSimpleError(err)
		os.Exit(-1)
	}
}
