package worker

import (
	"context"
	"fmt"
	"io"

	"github.com/ohsu-comp-bio/funnel/events"
)

type ContainerEngine interface {
	Run(ctx context.Context) error

	Stop() error

	Inspect(ctx context.Context) (ContainerConfig, error)

	GetImage() string

	GetIO() (io.Reader, io.Writer, io.Writer)

	SetIO(stdin io.Reader, stdout io.Writer, stderr io.Writer)
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

type ContainerEngineFactory struct{}

func (f *ContainerEngineFactory) NewContainerEngine(containerType string, containerConfig ContainerConfig) (ContainerEngine, error) {
	switch containerType {
	// case "docker":
	// 	return NewDockerEngine(containerConfig)
	case "exadocker":
		return NewExadockerEngine(containerConfig)
	default:
		return nil, fmt.Errorf("unsupported container type: %s", containerType)
	}
}

// func NewDockerEngine(config ContainerConfig) (ContainerEngine, error) {
// 	return Docker{
// 		ContainerConfig: config,
// 	}, nil
// }

func NewExadockerEngine(config ContainerConfig) (ContainerEngine, error) {
	return &Exadocker{
		ContainerConfig: config,
	}, nil
}
