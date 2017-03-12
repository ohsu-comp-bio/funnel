package worker

import (
	pscpu "github.com/shirou/gopsutil/cpu"
	psmem "github.com/shirou/gopsutil/mem"
	"net"
	"os"
	"os/exec"
	"path"
	"syscall"
	"tes/config"
	pbr "tes/server/proto"
)

const headerSize = int64(102400)

// exists returns whether the given file or directory exists or not
func exists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func ensureDir(p string) error {
	e, err := exists(p)
	if err != nil {
		return err
	}
	if !e {
		// TODO configurable mode?
		_ = syscall.Umask(0000)
		err := os.MkdirAll(p, 0777)
		if err != nil {
			return err
		}
	}
	return nil
}

func ensurePath(p string) error {
	dir := path.Dir(p)
	return ensureDir(dir)
}

func ensureFile(p string, class string) error {
	err := ensurePath(p)
	if err != nil {
		return err
	}
	if class == "File" {
		f, err := os.Create(p)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

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
func getExitCode(err error) int32 {
	if err != nil {
		if exiterr, exitOk := err.(*exec.ExitError); exitOk {
			if status, statusOk := exiterr.Sys().(syscall.WaitStatus); statusOk {
				return int32(status.ExitStatus())
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
func detectResources(conf *pbr.Resources) *pbr.Resources {
	res := &pbr.Resources{
		Cpus: conf.GetCpus(),
		Ram:  conf.GetRam(),
		Disk: conf.GetDisk(),
	}
	cpuinfo, _ := pscpu.Info()
	vmeminfo, _ := psmem.VirtualMemory()

	if conf.GetCpus() == 0 {
		// TODO is cores the best metric? with hyperthreading,
		//      runtime.NumCPU() and pscpu.Counts() return 8
		//      on my 4-core mac laptop
		for _, cpu := range cpuinfo {
			res.Cpus += uint32(cpu.Cores)
		}
	}

	if conf.GetRam() == 0.0 {
		res.Ram = float64(vmeminfo.Total) /
			float64(1024) / float64(1024) / float64(1024)
	}

	return res
}

// NoopJobRunner is useful during testing for creating a worker with a JobRunner
// that doesn't do anything.
func NoopJobRunner(l JobControl, c config.Worker, j *pbr.JobWrapper, u logUpdateChan) {
}
