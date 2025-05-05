package shared

import (
	"context"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/plugins/proto"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// GRPCClient is an implementation of KV that talks over RPC.
type GRPCClient struct{ client proto.AuthorizeClient }

func (m *GRPCClient) Get(headers map[string]*proto.StringList, params map[string]string, config *config.Config, task *tes.Task) (*proto.GetResponse, error) {
	resp, err := m.client.Get(context.Background(), &proto.GetRequest{
		Headers: headers,
		Params:  params,
		Config:  config,
		Task:    task,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

type GRPCServer struct {
	// This is the real implementation
	Impl Authorize
}

func (m *GRPCServer) Get(
	ctx context.Context,
	req *proto.GetRequest) (*proto.GetResponse, error) {
	v, err := m.Impl.Get(req.Params, req.Headers, req.Config, req.Task)
	return &proto.GetResponse{Config: v.Config,
		Code:    v.Code,
		Message: v.Message,
		Task:    v.Task}, err
}
