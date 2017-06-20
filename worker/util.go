package worker

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	pscpu "github.com/shirou/gopsutil/cpu"
	psdisk "github.com/shirou/gopsutil/disk"
	psmem "github.com/shirou/gopsutil/mem"
	"math"
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
			log.Info("Could not determine exit code. Using default -999", "err", err)
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

// recover from panic and call "cb" with an error value.
func handlePanic(cb func(error)) {
	if r := recover(); r != nil {
		if e, ok := r.(error); ok {
			cb(e)
		} else {
			cb(fmt.Errorf("Unknown task runner panic: %+v", r))
		}
	}
}

// helper aims to simplify the error and context checking in the runner code.
type helper struct {
	syserr       error
	execerr      error
	taskCanceled bool
	ctx          context.Context
}

func (h *helper) ok() bool {
	if h.ctx != nil {
		// Check if the context is done, but don't block waiting on it.
		select {
		case <-h.ctx.Done():
			h.syserr = h.ctx.Err()
		default:
		}
	}
	return h.syserr == nil && h.execerr == nil
}
