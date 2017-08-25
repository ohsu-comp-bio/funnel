package scheduler

import (
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/ohsu-comp-bio/funnel/config"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	pscpu "github.com/shirou/gopsutil/cpu"
	psdisk "github.com/shirou/gopsutil/disk"
	psmem "github.com/shirou/gopsutil/mem"
	"math"
	"os"
	"path"
	"path/filepath"
	"text/template"
	"time"
)

// DetectBinaryPath detects the path to the "funnel" binary
func DetectBinaryPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("Failed to detect path of funnel binary")
	}
	return path, err
}

// GenNodeID returns a UUID string.
func GenNodeID(prefix string) string {
	u, _ := uuid.NewV4()
	return fmt.Sprintf("%s-node-%s", prefix, u.String())
}

// detectResources helps determine the amount of resources to report.
// Resources are determined by inspecting the host, but they
// can be overridden by config.
func detectResources(conf config.Node) (pbs.Resources, error) {
	res := pbs.Resources{
		Cpus:   conf.Resources.Cpus,
		RamGb:  conf.Resources.RamGb,
		DiskGb: conf.Resources.DiskGb,
	}

	cpuinfo, err := pscpu.Info()
	if err != nil {
		return res, fmt.Errorf("can't detect cpu cores: %s", err)
	}
	vmeminfo, err := psmem.VirtualMemory()
	if err != nil {
		return res, fmt.Errorf("can't detect memory: %s", err)
	}
	diskinfo, err := psdisk.Usage(conf.WorkDir)
	if err != nil {
		return res, fmt.Errorf("can't detect available disk: %s", err)
	}

	if conf.Resources.Cpus == 0 {
		// TODO is cores the best metric? with hyperthreading,
		//      runtime.NumCPU() and pscpu.Counts() return 8
		//      on my 4-core mac laptop
		for _, cpu := range cpuinfo {
			res.Cpus += uint32(cpu.Cores)
		}
	}

	gb := math.Pow(1000, 3)
	if conf.Resources.RamGb == 0.0 {
		res.RamGb = float64(vmeminfo.Total) / float64(gb)
	}

	if conf.Resources.DiskGb == 0.0 {
		res.DiskGb = float64(diskinfo.Free) / float64(gb)
	}

	return res, nil
}

// SetupSingleTaskNode creates a node per task
func SetupSingleTaskNode(prefix string, c config.Node, t *tes.Task) *Offer {
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

	n := &pbs.Node{
		Id: prefix + t.Id,
		Resources: &pbs.Resources{
			Cpus:   cpus,
			RamGb:  ram,
			DiskGb: disk,
		},
	}

	n.Metadata = map[string]string{"project": project}

	if len(tzones) >= 1 {
		// TODO figure out zone mapping when len(tzones) > 1
		n.Zone = tzones[0]
	}

	return NewOffer(n, t, Scores{})
}

// SetupTemplatedHPCNode sets up a node in a HPC environment with a shared
// file system. It generates a submission file based on a template for
// schedulers such as SLURM, HTCondor, SGE, PBS/Torque, etc
func SetupTemplatedHPCNode(name string, tpl string, conf config.Config, n *pbs.Node) (string, error) {
	var err error

	nconf := conf.Scheduler.Node

	// TODO document that these working dirs need manual cleanup
	workdir := path.Join(nconf.WorkDir, n.Id)
	workdir, _ = filepath.Abs(workdir)
	err = util.EnsureDir(workdir)
	if err != nil {
		return "", err
	}

	nconf.ID = n.Id
	nconf.Timeout = 5 * time.Second
	nconf.Resources.Cpus = n.Resources.Cpus
	nconf.Resources.RamGb = n.Resources.RamGb
	nconf.Resources.DiskGb = n.Resources.DiskGb

	conf.Scheduler.Node = nconf
	confPath := path.Join(workdir, "node.conf.yml")
	conf.ToYamlFile(confPath)

	binaryPath, err := DetectBinaryPath()
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
		"NodeId":     n.Id,
		"Executable": binaryPath,
		"Config":     confPath,
		"WorkDir":    workdir,
		"Cpus":       int(n.Resources.Cpus),
		"RamGb":      n.Resources.RamGb,
		"DiskGb":     n.Resources.DiskGb,
		"Zone":       n.Zone,
		"Project":    n.Metadata["project"],
	})
	if err != nil {
		return "", err
	}
	f.Close()

	return submitPath, nil
}
