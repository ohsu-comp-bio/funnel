package plugins

import (
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-plugin"
)

// Define the Authorize type
type AuthorizeImpl struct{}

func (a AuthorizeImpl) Get(user string) ([]byte, error) {
	if user == "" {
		return nil, fmt.Errorf("user is required (e.g. ./authorize <user>)")
	}
	// Currently hardcoding the endpoint of the token service
	// TODO: This should be made configurabl
	resp, err := http.Get("http://localhost:8080/token?user=" + user)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}

	return body, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		Plugins: map[string]plugin.Plugin{
			"authorize": &AuthorizePlugin{Impl: &AuthorizeImpl{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
