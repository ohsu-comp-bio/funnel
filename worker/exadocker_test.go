package worker

import (
	"bytes"
	"context"
	"testing"

	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/stretchr/testify/assert"
)

func TestExadockerRun(t *testing.T) {
	var buf bytes.Buffer

	containerConfig := ContainerConfig{
		Image:   "ubuntu",
		Driver:  []string{"docker"},
		Command: []string{"echo", "hello"},
		Stdout:  &buf,
		Event:   events.NewExecutorWriter("task-1", 0, 0, events.Noop{}),
	}

	f := ContainerEngineFactory{}
	exadocker, err := f.NewContainerEngine("exadocker", containerConfig)

	err = exadocker.Run(context.Background())
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "hello")
}
