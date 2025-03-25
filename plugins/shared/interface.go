// 'shared' package contains shared data between the host and plugins.
package shared

import (
	"context"
	"net/rpc"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/plugins/proto"
	"google.golang.org/grpc"
)

// Define a struct that matches the expected JSON response
type Response struct {
	Code    int            `json:"code,omitempty"`
	Message string         `json:"message,omitempty"`
	Config  *config.Config `json:"config,omitempty"`
}

// Handshake is a common handshake that is shared by plugin and host.
var Handshake = plugin.HandshakeConfig{
	// This isn't required when using VersionedPlugins
	ProtocolVersion:  1,
	MagicCookieKey:   "AUTHORIZE_PLUGIN",
	MagicCookieValue: "authorize",
}

// Create an hclog.Logger
var Logger = hclog.New(&hclog.LoggerOptions{
	Name:  "plugin",
	Level: hclog.Trace,
})

// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	"authorize_grpc": &AuthorizeGRPCPlugin{},
	"authorize":      &AuthorizePlugin{},
}

// Authorize is the interface that we're exposing as a plugin.
type Authorize interface {
	Get(user string, host string) ([]byte, error)
}

// This is the implementation of plugin.Plugin so we can serve/consume this.
type AuthorizePlugin struct {
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Authorize
}

func (p *AuthorizePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &RPCServer{Impl: p.Impl}, nil
}

func (*AuthorizePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &RPCClient{client: c}, nil
}

// This is the implementation of plugin.GRPCPlugin so we can serve/consume this.
type AuthorizeGRPCPlugin struct {
	// GRPCPlugin must still implement the Plugin interface
	plugin.Plugin
	// Concrete implementation, written in Go. This is only used for plugins
	// that are written in Go.
	Impl Authorize
}

func (p *AuthorizeGRPCPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterAuthorizeServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

func (p *AuthorizeGRPCPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewAuthorizeClient(c)}, nil
}
