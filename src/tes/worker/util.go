package worker

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path"
	"syscall"
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
	return "", fmt.Errorf("Error no network connection")
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
