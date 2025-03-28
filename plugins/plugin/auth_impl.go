package plugin

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-plugin"
	"github.com/ohsu-comp-bio/funnel/plugins/shared"
)

// Here is a real implementation of Authorize that retrieves a "Secret" value for a user
type Authorize struct{}

func (Authorize) Get(user string, host string) ([]byte, error) {
	if user == "" {
		return nil, fmt.Errorf("user is required (e.g. ./authorize <USER> <HOST>)")
	}

	shared.Logger.Info("Get", "user", user, "host", host)
	resp, err := http.Get(host + user)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	shared.Logger.Info("Response", "status", resp.Status)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	shared.Logger.Info("Response", "body", string(body))
	return body, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: shared.Handshake,
		Plugins: map[string]plugin.Plugin{
			"authorize": &shared.AuthorizePlugin{Impl: &Authorize{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
