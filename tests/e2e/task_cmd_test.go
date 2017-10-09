package e2e

import (
	"bytes"
	"github.com/ohsu-comp-bio/funnel/cmd/run"
	"github.com/ohsu-comp-bio/funnel/cmd/task"
	"os"
	"strings"
	"testing"
)

// Test that the task commands are able to set the funnel server
// via a FUNNEL_SERVER environment variable.
func TestTaskCreateCmdServerEnvVar(t *testing.T) {
	var out bytes.Buffer
	os.Setenv("FUNNEL_SERVER", fun.Conf.Server.HTTPAddress())
	task.Cmd.SetArgs([]string{"create", "hello-world.json"})
	task.Cmd.SetOutput(&out)
	task.Cmd.Execute()

	t.Log(out.String())

	if strings.Contains(out.String(), "Error") {
		t.Errorf("unexpected error: %s", out.String())
	}
}

// Test that the task commands are able to set the funnel server
// via a FUNNEL_SERVER environment variable.
func TestTaskGetCmdServerEnvVar(t *testing.T) {

	id := fun.Run(`
    --sh 'echo hello world'
  `)

	var out bytes.Buffer
	os.Setenv("FUNNEL_SERVER", fun.Conf.Server.HTTPAddress())
	task.Cmd.SetArgs([]string{"get", id})
	task.Cmd.SetOutput(&out)
	task.Cmd.Execute()

	t.Log(out.String())

	if strings.Contains(out.String(), "Error") {
		t.Errorf("unexpected error: %s", out.String())
	}
}

// Test that the task commands are able to set the funnel server
// via a FUNNEL_SERVER environment variable.
func TestTaskListCmdServerEnvVar(t *testing.T) {
	var out bytes.Buffer
	os.Setenv("FUNNEL_SERVER", fun.Conf.Server.HTTPAddress())
	task.Cmd.SetArgs([]string{"list"})
	task.Cmd.SetOutput(&out)
	task.Cmd.Execute()

	t.Log(out.String())

	if strings.Contains(out.String(), "Error") {
		t.Errorf("unexpected error: %s", out.String())
	}
}

// Test that the run command is able to set the funnel server
// via a FUNNEL_SERVER environment variable.
func TestRunCmdServerEnvVar(t *testing.T) {
	var out bytes.Buffer
	os.Setenv("FUNNEL_SERVER", fun.Conf.Server.HTTPAddress())
	run.Cmd.SetArgs([]string{"echo hi"})
	run.Cmd.SetOutput(&out)
	run.Cmd.Execute()

	t.Log(out.String())

	if strings.Contains(out.String(), "Error") {
		t.Errorf("unexpected error: %s", out.String())
	}
}
