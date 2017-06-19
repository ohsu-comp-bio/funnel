package util

import (
	"encoding/base64"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// PerRPCPassword returns a new gRPC DialOption which includes a basic auth.
// password header in each RPC request.
func PerRPCPassword(password string) grpc.DialOption {
	return grpc.WithPerRPCCredentials(&loginCreds{
		Password: password,
	})
}

type loginCreds struct {
	Password string
}

func (c *loginCreds) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	v := base64.StdEncoding.EncodeToString([]byte("funnel:" + c.Password))

	return map[string]string{
		"Authorization": "Basic " + v,
	}, nil
}

func (c *loginCreds) RequireTransportSecurity() bool {
	return false
}
