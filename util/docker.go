package util

import (
	"context"
	"github.com/docker/docker/client"
	"os"
	"regexp"
)

// NewDockerClient returns a new docker client. This util handles
// working around some client/server API version mismatch issues.
func NewDockerClient() (*client.Client, error) {
	dclient, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	// If the api version is not set test if the client can communicate with the
	// server; if not infer API version from error message and inform the client
	// to use that version for future communication
	if os.Getenv("DOCKER_API_VERSION") == "" {
		_, err := dclient.ServerVersion(context.Background())
		if err != nil {
			re := regexp.MustCompile(`([0-9\.]+)`)
			version := re.FindAllString(err.Error(), -1)
			// Error message example:
			//   Error getting metadata for container: Error response from daemon: client is newer than server (client API version: 1.26, server API version: 1.24)
			os.Setenv("DOCKER_API_VERSION", version[1])
			return NewDockerClient()
		}
	}
	return dclient, nil
}
