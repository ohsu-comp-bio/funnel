package worker

import (
	"context"
	"fmt"
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
func getExitCode(err error) (code int, ok bool) {
	if err == nil {
		// The error is nil, the command returned successfully, so exit status is 0.
		ok = true
		return
	}

	if exiterr, exitOk := err.(*exec.ExitError); exitOk {
		if status, statusOk := exiterr.Sys().(syscall.WaitStatus); statusOk {
			code = status.ExitStatus()
			ok = true
			return
		}
	}

	// Could not determine exit status
	return
}

// recover from panic and call "cb" with an error value.
func handlePanic(cb func(error)) {
	if r := recover(); r != nil {
		if e, ok := r.(error); ok {
			cb(e)
		} else {
			cb(fmt.Errorf("Unknown worker panic: %+v", r))
		}
	}
}

// helper aims to simplify the error and context checking in the worker code.
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
