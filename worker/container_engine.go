package worker

import (
	"context"
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
	Command         string
	Volumes         []Volume
	Workdir         string
	RemoveContainer bool
	Env             map[string]string
	Stdin           io.Reader
	Stdout          io.Writer
	Stderr          io.Writer
	Event           *events.ExecutorWriter
	DriverCommand   string
	RunCommand      string // template string
	PullCommand     string // template string
	StopCommand     string // template string
}

type ContainerVersion struct {
	Client string
	Server string
}

type ContainerEngineFactory struct{}

func (f *ContainerEngineFactory) NewContainerEngine(containerConfig ContainerConfig) (ContainerEngine, error) {
	return NewDockerEngine(containerConfig)
}

func NewDockerEngine(config ContainerConfig) (ContainerEngine, error) {
	return &Docker{
		ContainerConfig: config,
	}, nil
}
