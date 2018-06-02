package server

import (
	"encoding/base64"
	"strings"

	"github.com/ohsu-comp-bio/funnel/config"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// Return a new interceptor function that authorizes RPCs
// using a password stored in the config.
func newAuthInterceptor(creds []config.BasicCredential) grpc.UnaryServerInterceptor {

	// Return a function that is the interceptor.
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {
		var authorized bool
		var err error
		for _, cred := range creds {
			err = authorize(ctx, cred.User, cred.Password)
			if err == nil {
				authorized = true
			}
		}
		if len(creds) == 0 {
			authorized = true
		}
		if !authorized {
			return nil, err
		}
		return handler(ctx, req)
	}
}

// Check the context's metadata for the configured server/API password.
func authorize(ctx context.Context, user, password string) error {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if len(md["authorization"]) > 0 {
			raw := md["authorization"][0]
			requser, reqpass, ok := parseBasicAuth(raw)
			if ok {
				if requser == user && reqpass == password {
					return nil
				}
				return grpc.Errorf(codes.PermissionDenied, "")
			}
		}
	}

	return grpc.Errorf(codes.Unauthenticated, "")
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
//
// Taken from Go core: https://golang.org/src/net/http/request.go?s=27379:27445#L828
func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = "Basic "

	if !strings.HasPrefix(auth, prefix) {
		return
	}

	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])

	if err != nil {
		return
	}

	cs := string(c)
	s := strings.IndexByte(cs, ':')

	if s < 0 {
		return
	}

	return cs[:s], cs[s+1:], true
}
