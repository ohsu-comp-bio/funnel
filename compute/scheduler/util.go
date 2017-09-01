package scheduler

import (
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/ohsu-comp-bio/funnel/config"
	pbs "github.com/ohsu-comp-bio/funnel/proto/scheduler"
	pscpu "github.com/shirou/gopsutil/cpu"
	psdisk "github.com/shirou/gopsutil/disk"
	psmem "github.com/shirou/gopsutil/mem"
	"math"
)

// GenNodeID returns a UUID string.
func GenNodeID(prefix string) string {
	u, _ := uuid.NewV4()
	return fmt.Sprintf("%s-node-%s", prefix, u.String())
}

// detectResources helps determine the amount of resources to report.
// Resources are determined by inspecting the host, but they
// can be overridden by config.
func detectResources(conf config.Node) pbs.Resources {
	res := pbs.Resources{
		Cpus:   conf.Resources.Cpus,
		RamGb:  conf.Resources.RamGb,
		DiskGb: conf.Resources.DiskGb,
	}

	cpuinfo, err := pscpu.Info()
	if err != nil {
		log.Error("Error detecting cpu cores", err)
		return res
	}
	vmeminfo, err := psmem.VirtualMemory()
	if err != nil {
		log.Error("Error detecting memory", err)
		return res
	}
	diskinfo, err := psdisk.Usage(conf.WorkDir)
	if err != nil {
		log.Error("Error detecting available disk", err)
		return res
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

	return res
}
