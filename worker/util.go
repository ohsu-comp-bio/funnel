package worker

import (
	"github.com/ohsu-comp-bio/funnel/config"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	pscpu "github.com/shirou/gopsutil/cpu"
	psdisk "github.com/shirou/gopsutil/disk"
	psmem "github.com/shirou/gopsutil/mem"
	"net"
	"os/exec"
	"syscall"
)

func externalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down

		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface

		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err

		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", nil
}

// getExitCode gets the exit status (i.e. exit code) from the result of an executed command.
// The exit code is zero if the command completed without error.
func getExitCode(err error) int {
	if err != nil {
		if exiterr, exitOk := err.(*exec.ExitError); exitOk {
			if status, statusOk := exiterr.Sys().(syscall.WaitStatus); statusOk {
				return status.ExitStatus()
			}
		} else {
			log.Info("Could not determine exit code. Using default -999")
			return -999
		}
	}
	// The error is nil, the command returned successfully, so exit status is 0.
	return 0
}

// detectResources helps determine the amount of resources to report.
// Resources are determined by inspecting the host, but they
// can be overridden by config.
func detectResources(conf config.Worker) config.Resources {
	res := config.Resources{
		Cpus:   conf.Resources.Cpus,
		RamGb:  conf.Resources.RamGb,
		DiskGb: conf.Resources.DiskGb,
	}
	cpuinfo, _ := pscpu.Info()
	vmeminfo, _ := psmem.VirtualMemory()
	diskinfo, _ := psdisk.Usage(conf.WorkDir)

	if conf.Resources.Cpus == 0 {
		// TODO is cores the best metric? with hyperthreading,
		//      runtime.NumCPU() and pscpu.Counts() return 8
		//      on my 4-core mac laptop
		for _, cpu := range cpuinfo {
			res.Cpus += uint32(cpu.Cores)
		}
	}

	if conf.Resources.RamGb == 0.0 {
		res.RamGb = float64(vmeminfo.Total) / float64(GB)
	}

	if conf.Resources.DiskGb == 0.0 {
		res.DiskGb = float64(diskinfo.Free) / float64(GB)
	}

	return res
}

// NoopTaskRunner is useful during testing for creating a worker with a TaskRunner
// that doesn't do anything.
func NoopTaskRunner(l TaskControl, c config.Worker, j *pbf.TaskWrapper) {
}

const (
	B  = 1
	KB = 1024 * B
	MB = 1024 * KB
	GB = 1024 * MB
)
