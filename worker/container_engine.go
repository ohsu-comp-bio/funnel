package worker

import (
	"context"
	"fmt"
	"io"

	"github.com/ohsu-comp-bio/funnel/events"
)

type ContainerEngine interface {
	// Run runs the container.
	Run(ctx context.Context) error

	// Stop stops the container.
	Stop() error

	// Inspect returns the container configuration.
	InspectContainer(ctx context.Context) ContainerConfig

	// GetImage returns the image name for the container.
	GetImage() string

	// GetIO returns the stdin, stdout, and stderr for the container.
	GetIO() (io.Reader, io.Writer, io.Writer)

	// SetIO sets the stdin, stdout, and stderr for the container.
	SetIO(stdin io.Reader, stdout io.Writer, stderr io.Writer)

	// SyncAPIVersions ensures that the client uses the same API version as the server.
	SyncAPIVersion() error
}

type ContainerConfig struct {
	Id              string
	Image           string
	Name            string
	Driver          []string
	Command         []string
	Volumes         []Volume
	Workdir         string
	ContainerName   string
	RemoveContainer bool
	Env             map[string]string
	Stdin           io.Reader
	Stdout          io.Writer
	Stderr          io.Writer
	Event           *events.ExecutorWriter
}

type ContainerVersion struct {
	Client string
	Server string
}

type ContainerEngineFactory struct{}

func (f *ContainerEngineFactory) NewContainerEngine(containerType string, containerConfig ContainerConfig) (ContainerEngine, error) {
	switch containerType {
	case "docker":
		return NewDockerEngine(containerConfig)
	case "exadocker":
		return NewExadockerEngine(containerConfig)
	default:
		return nil, fmt.Errorf("unsupported container type: %s", containerType)
	}
}

func NewDockerEngine(config ContainerConfig) (ContainerEngine, error) {
	return &Docker{
		ContainerConfig: config,
	}, nil
}

func NewExadockerEngine(config ContainerConfig) (ContainerEngine, error) {
	return &Exadocker{
		ContainerConfig: config,
	}, nil
}
