package util

import (
	"encoding/base64"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// DialOpts helps build a []grpc.DialOption
type DialOpts []grpc.DialOption

// Add adds the given option
func (d *DialOpts) Add(opt grpc.DialOption) {
	*d = append(*d, opt)
}

// TLS sets up transport security with the given certificate file.
// If cert == "", grpc.WithInsecure is used.
func (d *DialOpts) TLS(cert string) error {
	if cert != "" {
		creds, err := credentials.NewClientTLSFromFile(cert, "localhost")
		if err != nil {
			return err
		}
		d.Add(grpc.WithTransportCredentials(creds))
	} else {
		d.Add(grpc.WithInsecure())
	}
	return nil
}

// Password sets up a per-RPC basic auth. password.
// If p == "", so password is used.
func (d *DialOpts) Password(p string) {
	// TODO something needs to validate the config, to ensure
	// that if there is a server password, that TLS is setup.
	if p != "" {
		creds := loginCreds{p}
		d.Add(grpc.WithPerRPCCredentials(&creds))
	}
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
	return true
}
