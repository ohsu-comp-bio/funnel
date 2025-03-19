package plugins

import (
	"net/rpc"
)

// RPCClient is an implementation of Authorization that talks over RPC.
type RPCClient struct{ client *rpc.Client }

func (m *RPCClient) Get(user string) ([]byte, error) {
	var resp []byte
	err := m.client.Call("Plugin.Get", user, &resp)
	return resp, err
}

// Here is the RPC server that RPCClient talks to, conforming to
// the requirements of net/rpc
type RPCServer struct {
	// This is the real implementation
	Impl Authorize
}

func (m *RPCServer) Get(user string, resp *[]byte) error {
	v, err := m.Impl.Get(user)
	*resp = v
	return err
}
