package logger

import (
	"fmt"
	"google.golang.org/grpc/grpclog"
	"os"
)

func init() {
	ConfigureGRPC(2, NewLogger("grpc", DefaultConfig()))
}

// ConfigureGRPC configures the GRPC logger verboisty level.
//   All logs in transport package only go to verbose level 2.
//   All logs in other packages in grpc are logged in spite of the verbosity level.
func ConfigureGRPC(verbosity int, log *Logger) {
	grpclog.SetLoggerV2(&grpclogger{
		verbosity: verbosity,
		log:       log,
	})
}

// Configure the GRPC logger to use the global logrus configuration
type grpclogger struct {
	verbosity int
	log       *Logger
}

func (g *grpclogger) Info(args ...interface{}) {
	g.log.Info(fmt.Sprint(args))
}
func (g *grpclogger) Infoln(args ...interface{}) {
	g.log.Info(fmt.Sprint(args))
}
func (g *grpclogger) Infof(format string, args ...interface{}) {
	g.log.Info(fmt.Sprintf(format, args))
}
func (g *grpclogger) Warning(args ...interface{}) {
	g.log.Error(fmt.Sprint(args))
}
func (g *grpclogger) Warningln(args ...interface{}) {
	g.log.Error(fmt.Sprint(args))
}
func (g *grpclogger) Warningf(format string, args ...interface{}) {
	g.log.Error(fmt.Sprintf(format, args))
}
func (g *grpclogger) Error(args ...interface{}) {
	g.log.Error(fmt.Sprint(args))
}
func (g *grpclogger) Errorln(args ...interface{}) {
	g.log.Error(fmt.Sprint(args))
}
func (g *grpclogger) Errorf(format string, args ...interface{}) {
	g.log.Error(fmt.Sprintf(format, args))
}
func (g *grpclogger) Fatal(args ...interface{}) {
	g.log.Error(fmt.Sprint(args))
	os.Exit(1)
}
func (g *grpclogger) Fatalln(args ...interface{}) {
	g.log.Error(fmt.Sprint(args))
	os.Exit(1)
}
func (g *grpclogger) Fatalf(format string, args ...interface{}) {
	g.log.Error(fmt.Sprintf(format, args))
	os.Exit(1)
}
func (g *grpclogger) V(l int) bool {
	return g.verbosity >= l
}
