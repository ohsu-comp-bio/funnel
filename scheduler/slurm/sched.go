package slurm

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

var log = logger.New("slurm")

// prefix is a string prefixed to slurm worker IDs, so that slurm
// workers can be identified by ShouldStartWorker() below.
const prefix = "slurm-worker-"

// Plugin provides the SLURM scheduler backend plugin.
var Plugin = &scheduler.BackendPlugin{
	Name:   "slurm",
	Create: NewBackend,
}

// NewBackend returns a new SLURM Backend instance.
func NewBackend(conf config.Config) (scheduler.Backend, error) {
	return scheduler.Backend(&Backend{conf}), nil
}

// Backend represents the SLURM backend.
type Backend struct {
	conf config.Config
}

// Schedule schedules a task on the HTCondor queue and returns a corresponding Offer.
func (s *Backend) Schedule(t *tes.Task) *scheduler.Offer {
	log.Debug("Running slurm scheduler")

	disk := s.conf.Worker.Resources.DiskGb
	if disk == 0.0 {
		disk = t.GetResources().GetSizeGb()
	}

	cpus := s.conf.Worker.Resources.Cpus
	if cpus == 0 {
		cpus = t.GetResources().GetCpuCores()
	}

	ram := s.conf.Worker.Resources.RamGb
	if ram == 0.0 {
		ram = t.GetResources().GetRamGb()
	}

	w := &pbf.Worker{
		Id: prefix + t.Id,
		Resources: &pbf.Resources{
			Cpus:   cpus,
			RamGb:  ram,
			DiskGb: disk,
		},
	}
	return scheduler.NewOffer(w, t, scheduler.Scores{})
}

// ShouldStartWorker is part of the Scaler interface and returns true
// when the given worker needs to be started by Backend.StartWorker
func (s *Backend) ShouldStartWorker(w *pbf.Worker) bool {
	return strings.HasPrefix(w.Id, prefix) &&
		w.State == pbf.WorkerState_UNINITIALIZED
}

// StartWorker submits a task via "srun" to start a new worker.
func (s *Backend) StartWorker(w *pbf.Worker) error {
	log.Debug("Starting slurm worker")
	var err error

	// TODO document that these working dirs need manual cleanup
	workdir := path.Join(s.conf.WorkDir, w.Id)
	workdir, _ = filepath.Abs(workdir)
	err = os.MkdirAll(workdir, 0755)
	if err != nil {
		return err
	}

	wc := s.conf
	wc.Worker.ID = w.Id
	wc.Worker.Timeout = 5 * time.Second
	wc.Worker.Resources.Cpus = w.Resources.Cpus
	wc.Worker.Resources.RamGb = w.Resources.RamGb
	wc.Worker.Resources.DiskGb = w.Resources.DiskGb

	confPath := path.Join(workdir, "worker.conf.yml")
	wc.ToYamlFile(confPath)

	workerPath, err := scheduler.DetectWorkerPath()
	if err != nil {
		return err
	}

	submitPath := path.Join(workdir, "slurm.submit.sh")
	f, err := os.Create(submitPath)
	if err != nil {
		return err
	}

	submitTpl, err := template.New("slurm.submit.sh").Parse(`#!/bin/bash
#SBATCH --job-name {{.Name}}
#SBATCH --ntasks 1
#SBATCH --error {{.WorkDir}}/funnel-worker-stderr
#SBATCH --output {{.WorkDir}}/funnel-worker-stdout
{{.Resources}}

{{.ExecutableCmd}}
`)
	if err != nil {
		return err
	}

	err = submitTpl.Execute(f, map[string]string{
		"Name":          w.Id,
		"ExecutableCmd": fmt.Sprintf("%s --config %s", workerPath, confPath),
		"WorkDir":       workdir,
		"Resources":     resolveResourceRequest(int(w.Resources.Cpus), w.Resources.RamGb, w.Resources.DiskGb),
	})
	if err != nil {
		return err
	}
	f.Close()

	cmd := exec.Command(
		"sbatch",
		submitPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func resolveResourceRequest(cpus int, ram float64, disk float64) string {
	var resources = []string{}
	if cpus != 0 {
		resources = append(resources, fmt.Sprintf("#SBATCH --cpus-per-task %d", cpus))
	}
	if ram != 0.0 {
		resources = append(resources, fmt.Sprintf("#SBATCH --mem %.0fGB", ram))
	}
	// TODO: figure out if this is the right way to request disk space
	if disk != 0.0 {
		resources = append(resources, fmt.Sprintf("#SBATCH --tmp %.0fGB", disk))
	}
	return strings.Join(resources, "\n")
}
