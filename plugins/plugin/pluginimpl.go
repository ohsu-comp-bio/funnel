// Implementation of the plugin.Plugin interface for Auth.
package plugin

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
)

// AuthorizePlugin implements the plugin.Plugin interface to provide the RPC
// server or client back to the plugin machinery. The server side should
// proved the Impl field with a concrete implementation of the Auth
// interface.
type AuthorizePlugin struct {
	Impl Authorizer
}

func (p *AuthorizePlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &PluginServerRPC{
		Impl: p.Impl,
	}, nil
}

func (p *AuthorizePlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &PluginClientRPC{
		client: c,
	}, nil
}
