// RPC scaffolding for our server and client, using net/rpc.
//
// Eli Bendersky [https://eli.thegreenplace.net]
// This code is in the public domain.
package plugin

import (
	"log"
	"net/http"
	"net/rpc"

	"github.com/ohsu-comp-bio/funnel/auth"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// Types for RPC args/reply messages.

type HooksArgs struct{}

type HooksReply struct {
	Hooks []string
}

type AuthArgs struct {
	AuthHeader http.Header
	Task       *tes.Task
}

type AuthReply struct {
	Auth auth.Auth
	Err  error
}

// PluginServerRPC is used by plugins to map RPC calls from the clients to
// methods of the Auth interface.
type PluginServerRPC struct {
	Impl Authorizer
}

func (s *PluginServerRPC) Hooks(args HooksArgs, reply *HooksReply) error {
	reply.Hooks = s.Impl.Hooks()
	return nil
}

func (s *PluginServerRPC) Authorize(args AuthArgs, reply *AuthReply) error {
	reply.Auth, reply.Err = s.Impl.Authorize(args.AuthHeader, args.Task)
	return reply.Err
}

// PluginClientRPC is used by clients (main application) to translate the
// Authorize interface of plugins to RPC calls.
type PluginClientRPC struct {
	client *rpc.Client
}

func (c *PluginClientRPC) Hooks() []string {
	var reply HooksReply
	if err := c.client.Call("Plugin.Hooks", HooksArgs{}, &reply); err != nil {
		log.Printf("Error calling Plugin.Hooks: %v", err)
		return nil
	}
	return reply.Hooks
}

func (c *PluginClientRPC) Authorize(authHeader http.Header, task *tes.Task) (auth.Auth, error) {
	var reply AuthReply
	err := c.client.Call("Plugin.Authorize", AuthArgs{AuthHeader: authHeader, Task: task}, &reply)

	if err != nil {
		log.Printf("Error calling Plugin.Authorize: %v", err)
		return auth.Auth{}, err
	}

	return reply.Auth, reply.Err
}
