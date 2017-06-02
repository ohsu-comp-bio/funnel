package scheduler

import (
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/ohsu-comp-bio/funnel/config"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"os"
	"path"
	"path/filepath"
	"text/template"
	"time"
)

// DetectWorkerPath detects the path to the "funnel" binary
func DetectWorkerPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("Failed to detect path of funnel binary")
	}
	return path, err
}

// GenWorkerID returns a UUID string.
func GenWorkerID(prefix string) string {
	u, _ := uuid.NewV4()
	return fmt.Sprintf("%s-worker-%s", prefix, u.String())
}

// ScheduleSingleTaskWorker creates a worker per task
func ScheduleSingleTaskWorker(prefix string, c config.Worker, t *tes.Task) *Offer {
	disk := c.Resources.DiskGb
	if disk == 0.0 {
		disk = t.GetResources().GetSizeGb()
	}

	cpus := c.Resources.Cpus
	if cpus == 0 {
		cpus = t.GetResources().GetCpuCores()
	}

	ram := c.Resources.RamGb
	if ram == 0.0 {
		ram = t.GetResources().GetRamGb()
	}

	tzones := t.GetResources().GetZones()
	project := t.GetProject()

	w := &pbf.Worker{
		Id: prefix + t.Id,
		Resources: &pbf.Resources{
			Cpus:   cpus,
			RamGb:  ram,
			DiskGb: disk,
		},
	}

	w.Metadata = map[string]string{"project": project}

	if len(tzones) >= 1 {
		// TODO figure out zone mapping when len(tzones) > 1
		w.Zone = tzones[0]
	}

	return NewOffer(w, t, Scores{})
}

// SetupTemplatedHPCWorker sets up a worker in a HPC environment with a shared
// file system. It generates a submission file based on a template for
// schedulers such as SLURM, HTCondor, SGE, PBS/Torque, etc
func SetupTemplatedHPCWorker(name string, tpl string, conf config.Config, w *pbf.Worker) (string, error) {
	var err error

	// TODO document that these working dirs need manual cleanup
	workdir := path.Join(conf.Worker.WorkDir, w.Id)
	workdir, _ = filepath.Abs(workdir)
	err = util.EnsureDir(workdir)
	if err != nil {
		return "", err
	}

	wc := conf
	wc.Worker.ID = w.Id
	wc.Worker.Timeout = 5 * time.Second
	wc.Worker.Resources.Cpus = w.Resources.Cpus
	wc.Worker.Resources.RamGb = w.Resources.RamGb
	wc.Worker.Resources.DiskGb = w.Resources.DiskGb

	confPath := path.Join(workdir, "worker.conf.yml")
	wc.ToYamlFile(confPath)

	workerPath, err := DetectWorkerPath()
	if err != nil {
		return "", err
	}

	submitName := fmt.Sprintf("%s.submit", name)

	submitPath := path.Join(workdir, submitName)
	f, err := os.Create(submitPath)
	if err != nil {
		return "", err
	}

	submitTpl, err := template.New(submitName).Parse(tpl)
	if err != nil {
		return "", err
	}

	err = submitTpl.Execute(f, map[string]interface{}{
		"WorkerId":     w.Id,
		"Executable":   workerPath,
		"WorkerConfig": confPath,
		"WorkDir":      workdir,
		"Cpus":         int(w.Resources.Cpus),
		"RamGb":        w.Resources.RamGb,
		"DiskGb":       w.Resources.DiskGb,
		"Zone":         w.Zone,
		"Project":      w.Metadata["project"],
	})
	if err != nil {
		return "", err
	}
	f.Close()

	return submitPath, nil
}
