package e2e

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
	"testing"
	"time"
)

func TestHelloWorld(t *testing.T) {
	id := fun.Run(`
    --sh 'echo hello world'
  `)
	task := fun.Wait(id)

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("Missing stdout")
	}
}

func TestGetUnknownTask(t *testing.T) {
	var err error

	_, err = fun.HTTP.GetTask("nonexistent-task-id", "MINIMAL")
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 404") {
		log.Debug("error", err)
		t.Fatal("expected not found error")
	}

	_, err = fun.RPC.GetTask(
		context.Background(),
		&tes.GetTaskRequest{Id: "nonexistent-task-id", View: tes.TaskView_MINIMAL},
	)
	s, _ := status.FromError(err)
	if err == nil || s.Code() != codes.NotFound {
		log.Debug("error", err)
		t.Fatal("expected not found error")
	}
}

func TestGetTaskView(t *testing.T) {
	var err error
	var task *tes.Task

	fun.WriteFile("test_contents.txt", "hello world")

	id := fun.Run(`
    --sh 'echo hello world'
    --name 'foo'
    --contents in={{ .storage }}/test_contents.txt
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

	if len(task.Logs) != 1 {
		t.Fatal("expected TaskLog to be included in basic view")
	}

	if len(task.Logs[0].Logs) != 1 {
		t.Fatal("expected ExecutorLog to be included in basic view")
	}

	if task.Logs[0].Logs[0].Stdout != "" {
		t.Fatal("unexpected ExecutorLog stdout included in basic view")
	}

	if task.Inputs[0].Contents != "" {
		t.Fatal("unexpected TaskParameter contents included in basic view")
	}

	task = fun.GetView(id, tes.TaskView_FULL)

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("missing ExecutorLog stdout in full view")
	}

	if task.Inputs[0].Contents != "hello world" {
		t.Fatal("missing TaskParameter contents in full view")
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

	if len(task.Logs) != 1 {
		t.Fatal("expected TaskLog to be included in basic view")
	}

	if len(task.Logs[0].Logs) != 1 {
		t.Fatal("expected ExecutorLog to be included in basic view")
	}

	if task.Logs[0].Logs[0].Stdout != "" {
		t.Fatal("unexpected ExecutorLog stdout included in basic view")
	}

	if task.Inputs[0].Contents != "" {
		t.Fatal("unexpected TaskParameter contents included in basic view")
	}

	task, err = fun.HTTP.GetTask(id, "FULL")
	if err != nil {
		t.Fatal(err)
	}

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("missing ExecutorLog stdout in full view")
	}

	if task.Inputs[0].Contents != "hello world" {
		t.Fatal("missing TaskParameter contents in full view")
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
    --sh 'echo hello world'
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

	if len(task.Logs) != 1 {
		t.Fatal("expected TaskLog to be included in basic view")
	}

	if len(task.Logs[0].Logs) != 1 {
		t.Fatal("expected ExecutorLog to be included in basic view")
	}

	if task.Logs[0].Logs[0].Stdout != "" {
		t.Fatal("unexpected ExecutorLog stdout included in basic view")
	}

	tasks = fun.ListView(tes.TaskView_FULL)
	task = tasks[0]

	if task.Logs[0].Logs[0].Stdout != "hello world\n" {
		t.Fatal("Missing stdout in full view")
	}

	// test http proxy
	var r *tes.ListTasksResponse
	r, err = fun.HTTP.ListTasks(&tes.ListTasksRequest{
		View: tes.TaskView_MINIMAL,
	})
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

	r, err = fun.HTTP.ListTasks(&tes.ListTasksRequest{
		View: tes.TaskView_BASIC,
	})
	if err != nil {
		t.Fatal(err)
	}
	task = r.Tasks[0]

	if task.Name == "" {
		t.Fatal("expected task name to be included basic view")
	}

	if len(task.Logs) != 1 {
		t.Fatal("expected TaskLog to be included in basic view")
	}

	if len(task.Logs[0].Logs) != 1 {
		t.Fatal("expected ExecutorLog to be included in basic view")
	}

	if task.Logs[0].Logs[0].Stdout != "" {
		t.Fatal("unexpected ExecutorLog stdout included in basic view")
	}

	r, err = fun.HTTP.ListTasks(&tes.ListTasksRequest{
		View: tes.TaskView_FULL,
	})
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
    --sh "sh -c 'echo a; sleep 100'"
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
    --sh 'echo start'
    --sh 'sleep 10'
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
    --sh 'echo hello'
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
    --sh 'echo start'
    --sh 'sleep 1000'
    --sh 'echo never'
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

// Test canceling a task that doesn't exist
func TestCancelUnknownTask(t *testing.T) {
	var err error

	_, err = fun.HTTP.CancelTask("nonexistent-task-id")
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 404") {
		log.Debug("error", err)
		t.Fatal("expected not found error")
	}

	_, err = fun.RPC.CancelTask(
		context.Background(),
		&tes.CancelTaskRequest{Id: "nonexistent-task-id"},
	)
	s, _ := status.FromError(err)
	if err == nil || s.Code() != codes.NotFound {
		log.Debug("error", err)
		t.Fatal("expected not found error")
	}
}

// The task executor logs list should only include entries for steps that
// have been started or completed, i.e. steps that have yet to be started
// won't show up in Task.Logs[0].Logs
func TestExecutorLogLength(t *testing.T) {
	id := fun.Run(`
    --sh 'echo first'
    --sh 'sleep 10'
    --sh 'echo done'
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
    --sh 'echo step 1'
    --sh 'sleep 100'
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
	id := fun.Run(`--sh 'echo 1'`)
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

func TestPagination(t *testing.T) {
	c := DefaultConfig()
	c.Backend = "noop"
	f := NewFunnel(c)
	f.StartServer()
	ctx := context.Background()

	// Ensure database is empty
	r0, _ := f.RPC.ListTasks(ctx, &tes.ListTasksRequest{})
	if len(r0.Tasks) != 0 {
		t.Fatal("expected empty database")
	}

	for i := 0; i < 2050; i++ {
		f.Run(`--sh 'echo 1'`)
	}

	r1, _ := f.RPC.ListTasks(ctx, &tes.ListTasksRequest{})

	// Default page size is 256
	if len(r1.Tasks) != 256 {
		t.Error("wrong default page size")
	}

	r2, _ := f.RPC.ListTasks(ctx, &tes.ListTasksRequest{
		PageSize: 2,
	})

	// Minimum page size is 50
	if len(r2.Tasks) != 50 {
		t.Error("wrong minimum page size")
	}

	r3, _ := f.RPC.ListTasks(ctx, &tes.ListTasksRequest{
		PageSize: 5000,
	})

	if len(r3.Tasks) != 2048 {
		t.Error("wrong max page size")
	}

	r4, _ := f.RPC.ListTasks(ctx, &tes.ListTasksRequest{
		PageSize: 500,
	})

	if len(r4.Tasks) != 500 {
		t.Error("wrong requested page size")
	}

	if r4.NextPageToken == "" {
		t.Error("expected next page token")
	}

	// Get all pages
	var tasks []*tes.Task
	tasks = append(tasks, r4.Tasks...)
	for r4.NextPageToken != "" {
		r4, _ = f.RPC.ListTasks(ctx, &tes.ListTasksRequest{
			PageSize:  500,
			PageToken: r4.NextPageToken,
		})
		tasks = append(tasks, r4.Tasks...)
	}

	if len(tasks) != 2050 {
		log.Error("TASK COUNT", tasks)
		t.Error("unexpected task count")
	}
}

// Smaller test for debugging getting the full set of pages
func TestSmallPagination(t *testing.T) {
	c := DefaultConfig()
	c.Backend = "noop"
	f := NewFunnel(c)
	f.StartServer()
	ctx := context.Background()

	// Ensure database is empty
	r0, _ := f.RPC.ListTasks(ctx, &tes.ListTasksRequest{})
	if len(r0.Tasks) != 0 {
		t.Fatal("expected empty database")
	}

	for i := 0; i < 150; i++ {
		f.Run(`--sh 'echo 1'`)
	}

	r4, _ := f.RPC.ListTasks(ctx, &tes.ListTasksRequest{
		PageSize: 50,
	})

	// Get all pages
	var tasks []*tes.Task
	tasks = append(tasks, r4.Tasks...)
	for r4.NextPageToken != "" {
		r4, _ = f.RPC.ListTasks(ctx, &tes.ListTasksRequest{
			PageSize:  50,
			PageToken: r4.NextPageToken,
		})

		// Check a bug where the end of the last page was being included
		// in the start of the next page.
		if len(r4.Tasks) > 0 && r4.Tasks[0].Id == tasks[len(tasks)-1].Id {
			t.Error("Page start/end bug")
		}

		tasks = append(tasks, r4.Tasks...)
	}

	if len(tasks) != 150 {
		log.Error("TASK COUNT", len(tasks))
		t.Error("unexpected task count")
	}
}

func TestTaskError(t *testing.T) {
	id := fun.Run(`
    --sh "sh -c 'exit 1'"
  `)
	task := fun.Wait(id)

	if task.State != tes.State_ERROR {
		t.Fatal("Unexpected task state")
	}
}
