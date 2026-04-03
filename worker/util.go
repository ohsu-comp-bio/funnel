package worker

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime/debug"
	"strconv"
	"syscall"
)

// k8sExitCodeRegexp matches the exit code in Kubernetes executor error messages.
// e.g. "executor job test-123-0 failed with exit code 127 (Error): command not found"
var k8sExitCodeRegexp = regexp.MustCompile(`with exit code (\d+)`)

// getExitCode gets the exit status (i.e. exit code) from the result of an executed command.
// The exit code is zero if the command completed without error.
// Returns -999 if the exit code cannot be determined.
func getExitCode(err error) (int, error) {
	// The error is nil, the command returned successfully, so exit status is 0.
	if err == nil {
		return 0, nil
	}

	// Check for Kubernetes error type first
	if k8sErr, ok := err.(*K8sExecutorErr); ok {
		return k8sErr.ExitCode, nil
	}

	// Try to parse a Kubernetes-style error message string
	// e.g. "executor job test-123-0 failed with exit code 127 (Error): command not found"
	if matches := k8sExitCodeRegexp.FindStringSubmatch(err.Error()); len(matches) == 2 {
		if code, parseErr := strconv.Atoi(matches[1]); parseErr == nil {
			return code, nil
		}
	}

	if exiterr, exitOk := err.(*exec.ExitError); exitOk {
		if status, statusOk := exiterr.Sys().(syscall.WaitStatus); statusOk {
			return status.ExitStatus(), nil
		}
	}

	return -999, nil
}

// recover from panic and call "cb" with an error value.
func handlePanic(cb func(error)) {
	if r := recover(); r != nil {
		if e, ok := r.(error); ok {
			b := debug.Stack()
			cb(fmt.Errorf("panic: %s\n%s", e, string(b)))
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
