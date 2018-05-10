package builtin

import (
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/rs/xid"
	pscpu "github.com/shirou/gopsutil/cpu"
	psdisk "github.com/shirou/gopsutil/disk"
	psmem "github.com/shirou/gopsutil/mem"
)

// detectResources helps determine the amount of resources to report.
// Resources are determined by inspecting the host, but they
// can be overridden by config.
//
// Upon error, detectResources will return the resources given by the config
// with the error.
func detectResources(conf config.Node) (Resources, error) {
	res := Resources{
		Cpus:   conf.Resources.Cpus,
		RamGb:  conf.Resources.RamGb,
		DiskGb: conf.Resources.DiskGb,
	}

	cpuinfo, err := pscpu.Info()
	if err != nil {
		return res, fmt.Errorf("Error detecting cpu cores: %s", err)
	}
	vmeminfo, err := psmem.VirtualMemory()
	if err != nil {
		return res, fmt.Errorf("Error detecting memory: %s", err)
	}
	diskinfo, err := psdisk.Usage(conf.WorkDir)
	if err != nil {
		return res, fmt.Errorf("Error detecting available disk: %s", err)
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

func hostname() string {
	if name, err := os.Hostname(); err == nil {
		return name
	}
	return ""
}

// waitChan wraps sync.WaitGroup with a channel-based Wait()
// so it can be used in a select statement.
type waitChan struct {
	sync.WaitGroup
}

func (wc *waitChan) Wait() chan struct{} {
	out := make(chan struct{})
	go func() {
		wc.WaitGroup.Wait()
		close(out)
	}()
	return out
}

func genID() string {
	// chunk the ID, putting a separator near the end
	// to make it more readable.
	a := xid.New().String()
	b := a[:len(a)-3]
	c := a[len(a)-3:]
	return "node-" + b + "-" + c
}
