package rpc

import (
	"encoding/base64"
	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/ohsu-comp-bio/funnel/config"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"math"
	"math/rand"
	"time"
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

// Dial returns a new gRPC ClientConn with some default dial and call options set
func Dial(conf config.Server, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), conf.RPCClientTimeout)
	defer cancel()

	defaultOpts := []grpc.DialOption{
		grpc.WithInsecure(),
		PerRPCPassword(conf.Password),
	}
	opts = append(opts, defaultOpts...)
	opts = append(
		opts,
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(
			grpc_retry.WithMax(conf.RPCClientMaxRetries),
			grpc_retry.WithBackoff(exponentialBackoff),
		)),
	)

	return grpc.DialContext(ctx,
		conf.RPCAddress(),
		opts...,
	)
}

func exponentialBackoff(attempt uint) time.Duration {
	var initialBackoff = 5 * time.Second
	var maxBackoff = 1 * time.Minute
	var multiplier = 2.0
	var randomizationFactor = 0.5

	nextBackoff := jitter(float64(initialBackoff)*math.Pow(multiplier, float64(attempt)), randomizationFactor)

	if nextBackoff > float64(maxBackoff) {
		return time.Duration(jitter(float64(maxBackoff), randomizationFactor))
	}

	return time.Duration(nextBackoff)
}

func jitter(val float64, randomizationFactor float64) float64 {
	delta := randomizationFactor * val
	minInterval := val - delta
	maxInterval := val + delta

	// Get a random value from the range [minInterval, maxInterval].
	// The formula used below has a +1 because if the minInterval is 1 and the maxInterval is 3 then
	// we want a 33% chance for selecting either 1, 2 or 3.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return minInterval + (r.Float64() * (maxInterval - minInterval + 1))
}
