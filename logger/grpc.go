package logger

import (
	"fmt"
	"google.golang.org/grpc/grpclog"
)

func init() {
	// grpclog says to only call this from init(), so here we are
	grpclog.SetLogger(&grpclogger{})
}

// Wrap our logger to fit the grpc logger interface
type grpclogger struct {
}

func (g *grpclogger) Fatal(args ...interface{}) {
	global.Error("grpc", "msg", fmt.Sprint(args))
}
func (g *grpclogger) Fatalf(format string, args ...interface{}) {
	global.Error("grpc", "msg", fmt.Sprint(args))
}
func (g *grpclogger) Fatalln(args ...interface{}) {
	global.Error("grpc", "msg", fmt.Sprint(args))
}
func (g *grpclogger) Print(args ...interface{}) {
	global.Info("grpc", "msg", fmt.Sprint(args))
}
func (g *grpclogger) Printf(format string, args ...interface{}) {
	global.Info("grpc", "msg", fmt.Sprint(args))
}
func (g *grpclogger) Println(args ...interface{}) {
	global.Info("grpc", "msg", fmt.Sprint(args))
}
