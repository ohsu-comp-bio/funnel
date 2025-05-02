package shared

import (
	"fmt"
	"net/rpc"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/plugins/proto"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// RPCClient is an implementation of Authorize that talks over RPC.
type RPCClient struct {
	client *rpc.Client
}

type RPCServer struct {
	// This is the real implementation
	Impl Authorize
}

func (m *RPCClient) Get(params map[string]string, headers map[string]*proto.StringList, config *config.Config, task *tes.Task) ([]byte, error) {
	var resp []byte
	err := m.client.Call("Plugin.Get", &proto.GetRequest{
		Params:  params,
		Headers: headers,
		Config:  config,
		Task:    task,
	}, &resp)
	if err != nil {
		return nil, fmt.Errorf("PLUGIN RPC Get call failed: %w", err)
	}
	return resp, nil
}

func (m *RPCServer) Get(args *proto.GetRequest, resp *[]byte) error {
	// Call the implementation's Get method with the arguments
	v, err := m.Impl.Get(args.Params, args.Headers, args.Config, args.Task)
	*resp = v
	if err != nil {
		return fmt.Errorf("authorize implementation failed: %w", err)
	}
	return nil
}
