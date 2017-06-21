package gce

import (
	"github.com/ohsu-comp-bio/funnel/logger"
)

func init() {
	logger.Configure(logger.DebugConfig())
}
