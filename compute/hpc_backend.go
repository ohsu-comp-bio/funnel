package compute

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"text/template"
)

// NewHPCBackend returns a HPCBackend instance.
func NewHPCBackend(name string, submit string, conf config.Config, template string) *HPCBackend {
	return &HPCBackend{name, submit, conf, template}
}

// HPCBackend represents an HPCBackend such as HtCondor, Slurm, Grid Engine, etc.
type HPCBackend struct {
	name     string
	submit   string
	conf     config.Config
	template string
}

// Submit submits a task via "qsub"
func (b *HPCBackend) Submit(task *tes.Task) error {
	submitPath, err := b.setupTemplatedHPCSubmit(task)
	if err != nil {
		return err
	}

	cmd := exec.Command(b.submit, submitPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// setupTemplatedHPCSubmit sets up a task submission in a HPC environment with
// a shared file system. It generates a submission file based on a template for
// schedulers such as SLURM, HTCondor, SGE, PBS/Torque, etc.
func (b *HPCBackend) setupTemplatedHPCSubmit(task *tes.Task) (string, error) {
	var err error

	// TODO document that these working dirs need manual cleanup
	workdir := path.Join(b.conf.Worker.WorkDir, task.Id)
	workdir, _ = filepath.Abs(workdir)
	err = util.EnsureDir(workdir)
	if err != nil {
		return "", err
	}

	confPath := path.Join(workdir, "worker.conf.yml")
	b.conf.ToYamlFile(confPath)

	funnelPath, err := DetectFunnelBinaryPath()
	if err != nil {
		return "", err
	}

	submitName := fmt.Sprintf("%s.submit", b.name)

	submitPath := path.Join(workdir, submitName)
	f, err := os.Create(submitPath)
	if err != nil {
		return "", err
	}

	submitTpl, err := template.New(submitName).Parse(b.template)
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
