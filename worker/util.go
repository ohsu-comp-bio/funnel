package worker

import (
	"context"
	"fmt"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
)

// getExitCode gets the exit status (i.e. exit code) from the result of an executed command.
// The exit code is zero if the command completed without error.
func getExitCode(err error) int {
	fmt.Println("DEBUG: error", err)
	if err != nil {
		fmt.Printf("DEBUG: getExitCode err '%s'\n", err)
		if exiterr, exitOk := err.(*exec.ExitError); exitOk {
			if status, statusOk := exiterr.Sys().(syscall.WaitStatus); statusOk {
				return status.ExitStatus()
			}
		} else {
			// Try to extract exit code from Kubernetes error format
			// Format: "executor job <jobName> failed with exit code <code> ..."
			errStr := err.Error()
			fmt.Printf("DEBUG: Parsing error string: %s\n", errStr)

			if strings.Contains(errStr, "failed with exit code") {
				// Extract exit code from message like "failed with exit code 127"
				parts := strings.Split(errStr, "exit code ")
				if len(parts) > 1 {
					codeStr := strings.Fields(parts[1])[0]
					if code, parseErr := strconv.Atoi(codeStr); parseErr == nil {
						fmt.Printf("DEBUG: Parsed exit code: %d\n", code)
						return code
					}
				}
			}

			// Also try to match "failed with X failures" pattern
			if strings.Contains(errStr, "failed with") && strings.Contains(errStr, "failures") {
				// This is fallback case when container inspection fails
				// Return -1 to indicate unknown actual exit code
				fmt.Printf("DEBUG: Using fallback error pattern, returning -1\n")
				return -1
			}

			return -999 // could not get exit code, un-spec'd value that could break downstream
		}
	}
	// The error is nil, the command returned successfully, so exit status is 0.
	return 0
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
