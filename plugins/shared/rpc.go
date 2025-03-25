package shared

import (
	"fmt"
	"net/rpc"
)

// RPCClient is an implementation of Authorization that talks over RPC.
type RPCClient struct{ client *rpc.Client }

func (m *RPCClient) Get(user string, host string) ([]byte, error) {
	var resp []byte
	err := m.client.Call("Plugin.Get", []string{user, host}, &resp)
	return resp, err
}

// Here is the RPC server that RPCClient talks to, conforming to
// the requirements of net/rpc
type RPCServer struct {
	// This is the real implementation
	Impl Authorize
}

func (m *RPCServer) Get(args []string, resp *[]byte) error {
	if len(args) != 2 {
		return fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	user, host := args[0], args[1]
	v, err := m.Impl.Get(user, host)
	*resp = v
	return err
}
