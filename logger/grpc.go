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

// DisableGRPCLogger disables all grpc logging
func DisableGRPCLogger() {
	grpclog.SetLogger(&disabledlogger{})
}

// EnableGRPCLogger configures the GRPC logger to use the global logrus configuration
func EnableGRPCLogger() {
	grpclog.SetLogger(&grpclogger{})
}

// this is used in the tests so the RPC client doesn't spam logs while connecting
type disabledlogger struct {
}

func (g *disabledlogger) Fatal(args ...interface{}) {
}
func (g *disabledlogger) Fatalf(format string, args ...interface{}) {
}
func (g *disabledlogger) Fatalln(args ...interface{}) {
}
func (g *disabledlogger) Print(args ...interface{}) {
}
func (g *disabledlogger) Printf(format string, args ...interface{}) {
}
func (g *disabledlogger) Println(args ...interface{}) {
}
