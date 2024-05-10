package worker

import "context"

type Engine interface {
	Pull(ctx context.Context, image string) error

	Run(ctx context.Context, config ContainerConfig) error

	Stop(ctx context.Context, containerID string) error

	Inspect(ctx context.Context, containerID string) (ContainerInfo, error)
}

type ContainerConfig struct {
	Image   string
	Command []string
	Args    []string
	EnvVars map[string]string
	Volumes []Volume
}

type ContainerInfo struct {
	ID    string
	Name  string
	Image string
}
