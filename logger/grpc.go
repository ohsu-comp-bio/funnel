package logger

import (
	"fmt"
	"google.golang.org/grpc/grpclog"
)

var grpc = New("grpc")

func init() {
	// grpclog says to only call this from init(), so here we are
	grpclog.SetLogger(&grpclogger{})
}

// Wrap our logger to fit the grpc logger interface
type grpclogger struct {
}

func (g *grpclogger) Fatal(args ...interface{}) {
	grpc.Error(fmt.Sprint(args))
}
func (g *grpclogger) Fatalf(format string, args ...interface{}) {
	grpc.Error(fmt.Sprint(args))
}
func (g *grpclogger) Fatalln(args ...interface{}) {
	grpc.Error(fmt.Sprint(args))
}
func (g *grpclogger) Print(args ...interface{}) {
	grpc.Error(fmt.Sprint(args))
}
func (g *grpclogger) Printf(format string, args ...interface{}) {
	grpc.Error(fmt.Sprint(args))
}
func (g *grpclogger) Println(args ...interface{}) {
	grpc.Error(fmt.Sprint(args))
}
