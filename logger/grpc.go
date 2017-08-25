package logger

import (
	"fmt"
	"google.golang.org/grpc/grpclog"
	"os"
)

var log = New("grpc")

func init() {
	// grpclog says to only call this from init(), so here we are
	grpclog.SetLoggerV2(&grpclogger{2})
}

// SetGRPCLoggerVerbosity configures the GRPC logger verboisty level.
//   All logs in transport package only go to verbose level 2.
//   All logs in other packages in grpc are logged in spite of the verbosity level.
func SetGRPCLoggerVerbosity(verbosity int) {
	grpclog.SetLoggerV2(&grpclogger{verbosity})
}

// Configure the GRPC logger to use the global logrus configuration
type grpclogger struct {
	verbosity int
}

func (g *grpclogger) Info(args ...interface{}) {
	log.Info(fmt.Sprint(args))
}
func (g *grpclogger) Infoln(args ...interface{}) {
	log.Info(fmt.Sprint(args))
}
func (g *grpclogger) Infof(format string, args ...interface{}) {
	log.Info(fmt.Sprintf(format, args))
}
func (g *grpclogger) Warning(args ...interface{}) {
	log.Error(fmt.Sprint(args))
}
func (g *grpclogger) Warningln(args ...interface{}) {
	log.Error(fmt.Sprint(args))
}
func (g *grpclogger) Warningf(format string, args ...interface{}) {
	log.Error(fmt.Sprintf(format, args))
}
func (g *grpclogger) Error(args ...interface{}) {
	log.Error(fmt.Sprint(args))
}
func (g *grpclogger) Errorln(args ...interface{}) {
	log.Error(fmt.Sprint(args))
}
func (g *grpclogger) Errorf(format string, args ...interface{}) {
	log.Error(fmt.Sprintf(format, args))
}
func (g *grpclogger) Fatal(args ...interface{}) {
	log.Error(fmt.Sprint(args))
	os.Exit(1)
}
func (g *grpclogger) Fatalln(args ...interface{}) {
	log.Error(fmt.Sprint(args))
	os.Exit(1)
}
func (g *grpclogger) Fatalf(format string, args ...interface{}) {
	log.Error(fmt.Sprintf(format, args))
	os.Exit(1)
}
func (g *grpclogger) V(l int) bool {
	return g.verbosity >= l
}
