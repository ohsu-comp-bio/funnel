package compute

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"text/template"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/events"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

// HPCBackend represents an HPCBackend such as HtCondor, Slurm, Grid Engine, etc.
type HPCBackend struct {
	Name      string
	SubmitCmd string
	CancelCmd string
	Template  string
	Conf      config.Config
	Event     events.Writer
	Database  tes.ReadOnlyServer
	// ExtractID is responsible for extracting the task id from the response
	// returned by the SubmitCmd.
	ExtractID func(string) string
}

// WriteEvent writes an event to the compute backend.
// Currently, only TASK_CREATED is handled, which calls Submit.
func (b *HPCBackend) WriteEvent(ctx context.Context, ev *events.Event) error {
	switch ev.Type {
	case events.Type_TASK_CREATED:
		return b.Submit(ctx, ev.GetTask())

	case events.Type_TASK_STATE:
		if ev.GetState() == tes.State_CANCELED {
			return b.Cancel(ctx, ev.Id)
		}
	}
	return nil
}

// Submit submits a task via "qsub", "condor_submit", "sbatch", etc.
func (b *HPCBackend) Submit(ctx context.Context, task *tes.Task) error {
	submitPath, err := b.setupTemplatedHPCSubmit(task)
	if err != nil {
		return err
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(b.SubmitCmd, submitPath)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		b.Event.WriteEvent(ctx, events.NewState(task.Id, tes.SystemError))
		b.Event.WriteEvent(
			ctx,
			events.NewSystemLog(
				task.Id, 0, 0, "error",
				"error submitting task to "+b.Name,
				map[string]string{"error": err.Error(), "stderr": stderr.String(), "stdout": stdout.String()},
			),
		)
		return err
	}

	backendID := b.ExtractID(stdout.String())

	return b.Event.WriteEvent(
		ctx,
		events.NewMetadata(task.Id, 0, map[string]string{fmt.Sprintf("%s_id", b.Name): backendID}),
	)
}

// Cancel cancels a task via "qdel", "condor_rm", "scancel", etc.
func (b *HPCBackend) Cancel(ctx context.Context, taskID string) error {
	task, err := b.Database.GetTask(
		ctx, &tes.GetTaskRequest{Id: taskID, View: tes.TaskView_BASIC},
	)
	if err != nil {
		return err
	}

	// only cancel tasks in a QUEUED state
	if task.State != tes.State_QUEUED {
		return nil
	}

	backendID := getBackendTaskID(task, b.Name)
	if backendID == "" {
		return fmt.Errorf("no %s_id found in metadata for task %s", b.Name, taskID)
	}

	cmd := exec.Command(b.CancelCmd, backendID)
	return cmd.Run()
}

// setupTemplatedHPCSubmit sets up a task submission in a HPC environment with
// a shared file system. It generates a submission file based on a template for
// schedulers such as SLURM, HTCondor, SGE, PBS/Torque, etc.
func (b *HPCBackend) setupTemplatedHPCSubmit(task *tes.Task) (string, error) {
	var err error

	// TODO document that these working dirs need manual cleanup
	workdir := path.Join(b.Conf.Worker.WorkDir, task.Id)
	workdir, _ = filepath.Abs(workdir)
	err = fsutil.EnsureDir(workdir)
	if err != nil {
		return "", err
	}

	confPath := path.Join(workdir, "worker.conf.yml")
	config.ToYamlFile(b.Conf, confPath)

	funnelPath, err := DetectFunnelBinaryPath()
	if err != nil {
		return "", err
	}

	submitName := fmt.Sprintf("%s.submit", b.Name)

	submitPath := path.Join(workdir, submitName)
	f, err := os.Create(submitPath)
	if err != nil {
		return "", err
	}

	submitTpl, err := template.New(submitName).Parse(b.Template)
	if err != nil {
		return "", err
	}

	var zone string
	zones := task.Resources.GetZones()
	if zones != nil {
		zone = zones[0]
	}

	err = submitTpl.Execute(f, map[string]interface{}{
		"TaskId":     task.Id,
		"Executable": funnelPath,
		"Config":     confPath,
		"WorkDir":    workdir,
		"Cpus":       int(task.Resources.CpuCores),
		"RamGb":      task.Resources.RamGb,
		"DiskGb":     task.Resources.DiskGb,
		"Zone":       zone,
	})
	if err != nil {
		return "", err
	}
	f.Close()

	return submitPath, nil
}

func getBackendTaskID(task *tes.Task, backend string) string {
	logs := task.GetLogs()
	if len(logs) > 0 {
		metadata := logs[0].GetMetadata()
		return metadata[backend+"_id"]
	}
	return ""
}
