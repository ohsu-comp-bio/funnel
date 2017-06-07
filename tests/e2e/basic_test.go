package e2e

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	logger.ForceColors()
	fun.StartServer()
	e := m.Run()
	os.Exit(e)
}

func TestHelloWorld(t *testing.T) {
	id := fun.Run(`
    --cmd 'echo hello world'
  `)
	task := fun.Wait(id)

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("Missing stdout")
	}
}

func TestGetTaskView(t *testing.T) {
	var err error
	var task *tes.Task

	id := fun.Run(`
    --cmd 'echo hello world'
    --name 'foo'
  `)
	fun.Wait(id)

	task = fun.GetView(id, tes.TaskView_MINIMAL)

	if task.Id != id {
		t.Fatal("expected task ID in minimal view")
	}
	if task.State != tes.State_COMPLETE {
		t.Fatal("expected complete state")
	}
	if task.Name != "" {
		t.Fatal("unexpected task name included in minimal view")
	}
	if task.Logs != nil {
		t.Fatal("unexpected task logs included in minimal view")
	}

	task = fun.GetView(id, tes.TaskView_BASIC)

	if task.Name != "foo" {
		t.Fatal("expected task name to be included basic view")
	}

	if task.Logs != nil {
		t.Fatal("unexpected task logs included in basic view")
	}

	task = fun.GetView(id, tes.TaskView_FULL)

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("Missing stdout in full view")
	}

	// test http proxy
	task, err = fun.HTTP.GetTask(id, "MINIMAL")
	if err != nil {
		t.Fatal(err)
	}

	if task.Id != id {
		t.Fatal("expected task ID in minimal view")
	}
	if task.Name != "" {
		t.Fatal("unexpected task name included in minimal view")
	}
	if task.Logs != nil {
		t.Fatal("unexpected task logs included in minimal view")
	}

	task, err = fun.HTTP.GetTask(id, "BASIC")
	if err != nil {
		t.Fatal(err)
	}

	if task.Name != "foo" {
		t.Fatal("expected task name to be included basic view")
	}

	if task.Logs != nil {
		t.Fatal("unexpected task logs included in basic view")
	}

	task, err = fun.HTTP.GetTask(id, "FULL")
	if err != nil {
		t.Fatal(err)
	}

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("Missing stdout in full view")
	}
}

// TODO this is a bit hacky for now because we're reusing the same
//      server + DB for all the e2e tests, so ListTasks gets the
//      results of all of those. It works for the moment, but
//      should probably run against a clean environment.
func TestListTaskView(t *testing.T) {
	var tasks []*tes.Task
	var task *tes.Task
	var err error

	id := fun.Run(`
    --cmd 'echo hello world'
    --name 'foo'
  `)
	fun.Wait(id)

	tasks = fun.ListView(tes.TaskView_MINIMAL)
	task = tasks[0]

	if task.Id == "" {
		t.Fatal("expected task ID in minimal view")
	}
	if task.State == tes.State_UNKNOWN {
		t.Fatal("expected complete state")
	}
	if task.Name != "" {
		t.Fatal("unexpected task name included in minimal view")
	}
	if task.Logs != nil {
		t.Fatal("unexpected task logs included in minimal view")
	}

	tasks = fun.ListView(tes.TaskView_BASIC)
	task = tasks[0]

	if task.Name == "" {
		t.Fatal("expected task name to be included basic view")
	}

	if task.Logs != nil {
		t.Fatal("unexpected task logs included in basic view")
	}

	tasks = fun.ListView(tes.TaskView_FULL)
	task = tasks[0]

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("Missing stdout in full view")
	}

	// test http proxy
	var r *tes.ListTasksResponse
	r, err = fun.HTTP.ListTasks("MINIMAL")
	if err != nil {
		t.Fatal(err)
	}
	task = r.Tasks[0]

	if task.Id == "" {
		t.Fatal("expected task ID in minimal view")
	}
	if task.State == tes.State_UNKNOWN {
		t.Fatal("expected complete state")
	}
	if task.Name != "" {
		t.Fatal("unexpected task name included in minimal view")
	}
	if task.Logs != nil {
		t.Fatal("unexpected task logs included in minimal view")
	}

	r, err = fun.HTTP.ListTasks("BASIC")
	if err != nil {
		t.Fatal(err)
	}
	task = r.Tasks[0]

	if task.Name == "" {
		t.Fatal("expected task name to be included basic view")
	}

	if task.Logs != nil {
		t.Fatal("unexpected task logs included in basic view")
	}

	r, err = fun.HTTP.ListTasks("FULL")
	if err != nil {
		t.Fatal(err)
	}
	task = r.Tasks[0]

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("Missing stdout in full view")
	}
}

// Test that the streaming logs pick up a single character.
// This ensures that the streaming works even when a small
// amount of logs are written (which was once a bug).
func TestSingleCharLog(t *testing.T) {
	id := fun.Run(`
    --cmd "sh -c 'echo a; sleep 100'"
  `)
	fun.WaitForRunning(id)
	time.Sleep(time.Second * 2)
	task := fun.Get(id)
	if task.Logs[0].Logs[0].Stdout != "a\n" {
		t.Fatal("Missing logs")
	}
	fun.Cancel(id)
}

// Test that port mappings are being logged.
/* TODO need ports in funnel run
func TestPortLog(t *testing.T) {
  id := fun.Run(`
    --cmd 'echo start'
    --cmd 'sleep 10'
  `)
  fun.WaitForExec(id, 2)
  task := fun.Get(id)
  if task.Logs[0].Logs[0].Ports[0].Host != 5000 {
    t.Fatal("Unexpected port logs")
  }
}
*/

// Test that a completed task cannot change state.
func TestCompleteStateImmutable(t *testing.T) {
	id := fun.Run(`
    --cmd 'echo hello'
  `)
	fun.Wait(id)
	err := fun.Cancel(id)
	if err == nil {
		t.Fatal("expected error")
	}
	task := fun.Get(id)
	if task.State != tes.State_COMPLETE {
		t.Fatal("Unexpected state")
	}
}

// Test canceling a task
func TestCancel(t *testing.T) {
	id := fun.Run(`
    --cmd 'echo start'
    --cmd 'sleep 1000'
    --cmd 'echo never'
  `)
	fun.WaitForExec(id, 1)
	fun.Cancel(id)
	// TODO docker details and container ID are very Funnel specific
	//      how could we generalize this to be reusued as TES conformance?
	fun.WaitForDockerDestroy(id + "-0")
	task := fun.Get(id)
	if task.State != tes.State_CANCELED {
		t.Fatal("Unexpected state")
	}
}

// The task executor logs list should only include entries for steps that
// have been started or completed, i.e. steps that have yet to be started
// won't show up in Task.Logs[0].Logs
func TestExecutorLogLength(t *testing.T) {
	id := fun.Run(`
    --cmd 'echo first'
    --cmd 'sleep 10'
    --cmd 'echo done'
  `)
	fun.WaitForExec(id, 2)
	task := fun.Get(id)
	fun.Cancel(id)
	if len(task.Logs[0].Logs) != 2 {
		t.Fatal("Unexpected executor log count")
	}
}

// There was a bug + fix where the task was being marked complete after
// the first step completed, but the correct behavior is to mark the
// task complete after *all* steps have completed.
func TestMarkCompleteBug(t *testing.T) {
	id := fun.Run(`
    --cmd 'echo step 1'
    --cmd 'sleep 100'
  `)
	fun.WaitForRunning(id)
	fun.WaitForExec(id, 2)
	task := fun.Get(id)
	if task.State != tes.State_RUNNING {
		t.Fatal("Unexpected task state")
	}
	fun.Cancel(id)
}

func TestTaskStartEndTimeLogs(t *testing.T) {
	id := fun.Run(`--cmd 'echo 1'`)
	task := fun.Wait(id)
	if task.Logs[0].StartTime == "" {
		t.Fatal("missing task start time log")
	}
	if task.Logs[0].EndTime == "" {
		t.Fatal("missing task end time log")
	}
}

func TestOutputFileLog(t *testing.T) {
	dir := fun.Tempdir()

	id, _ := fun.RunTask(&tes.Task{
		Executors: []*tes.Executor{
			{
				ImageName: "alpine",
				Cmd: []string{
					"sh", "-c", "mkdir /tmp/outdir; echo fooo > /tmp/outdir/fooofile; echo ba > /tmp/outdir/bafile; echo bar > /tmp/barfile",
				},
			},
		},
		Outputs: []*tes.TaskParameter{
			{
				Url:  dir + "/outdir",
				Path: "/tmp/outdir",
				Type: tes.FileType_DIRECTORY,
			},
			{
				Url:  dir + "/barfile",
				Path: "/tmp/barfile",
			},
		},
	})

	task := fun.Wait(id)
	log.Debug("TEST", "task", task)

	out := task.Logs[0].Outputs

	if out[0].Url != dir+"/outdir/bafile" {
		t.Fatal("unexpected output url")
	}

	if out[1].Url != dir+"/outdir/fooofile" {
		t.Fatal("unexpected output url")
	}

	if out[2].Url != dir+"/barfile" {
		t.Fatal("unexpected output url")
	}

	if out[0].SizeBytes != 3 {
		t.Fatal("unexpected output size")
	}
	if out[1].SizeBytes != 5 {
		t.Fatal("unexpected output size")
	}
	if out[2].SizeBytes != 4 {
		t.Fatal("unexpected output size")
	}
}
