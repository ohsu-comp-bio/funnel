package e2e

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"testing"
	"time"
)

func TestHelloWorld(t *testing.T) {
	id := run(`
    --cmd 'echo hello world'
  `)
	task := wait(id)

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("Missing stdout")
	}
}

// Test that the streaming logs pick up a single character.
// This ensures that the streaming works even when a small
// amount of logs are written (which was once a bug).
func TestSingleCharLog(t *testing.T) {
	id := run(`
    --cmd "sh -c 'echo a; sleep 10'"
  `)
	time.Sleep(time.Second * 2)
	task := get(id)
	if task.Logs[0].Logs[0].Stdout != "a\n" {
		t.Fatal("Missing logs")
	}
	cancel(id)
}

// Test that port mappings are being logged.
/* TODO need ports in funnel run
func TestPortLog(t *testing.T) {
  id := run(`
    --cmd 'echo start'
    --cmd 'sleep 10'
  `)
  waitForExec(id, 2)
  task := get(id)
  if task.Logs[0].Logs[0].Ports[0].Host != 5000 {
    t.Fatal("Unexpected port logs")
  }
}
*/

// Test that a completed task cannot change state.
func TestCompleteStateImmutable(t *testing.T) {
	id := run(`
    --cmd 'echo hello'
  `)
	wait(id)
	err := cancel(id)
	if err == nil {
		t.Fatal("expected error")
	}
	task := get(id)
	if task.State != tes.State_COMPLETE {
		t.Fatal("Unexpected state")
	}
}

// Test canceling a task
func TestCancel(t *testing.T) {
	id := run(`
    --cmd 'echo start'
    --cmd 'sleep 1000'
    --cmd 'echo never'
  `)
	waitForExec(id, 1)
	cancel(id)
	// TODO docker details and container ID are very Funnel specific
	//      how could we generalize this to be reusued as TES conformance?
	waitForDockerDestroy(id + "-0")
	task := get(id)
	if task.State != tes.State_CANCELED {
		t.Fatal("Unexpected state")
	}
}

// The task executor logs list should only include entries for steps that
// have been started or completed, i.e. steps that have yet to be started
// won't show up in Task.Logs[0].Logs
func TestExecutorLogLength(t *testing.T) {
	id := run(`
    --cmd 'echo first'
    --cmd 'sleep 10'
    --cmd 'echo done'
  `)
	waitForExec(id, 2)
	task := get(id)
	cancel(id)
	if len(task.Logs[0].Logs) != 2 {
		t.Fatal("Unexpected executor log count")
	}
}

// There was a bug + fix where the task was being marked complete after
// the first step completed, but the correct behavior is to mark the
// task complete after *all* steps have completed.
func TestMarkCompleteBug(t *testing.T) {
	id := run(`
    --cmd 'echo step 1'
    --cmd 'sleep 100'
  `)
	waitForExec(id, 2)
	task := get(id)
	if task.State != tes.State_RUNNING {
		t.Fatal("Unexpected task state")
	}
}
