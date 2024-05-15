package worker

import (
	"context"
	"io"

	"github.com/ohsu-comp-bio/funnel/events"
)

type ContainerEngine interface {
	Run(ctx context.Context) error

	Stop() error

	Inspect(ctx context.Context) (ContainerInfo, error)
}

type ContainerCommand struct {
	Image           string
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

type ContainerInfo struct {
	Id    string
	Image string
	Name  string
}
