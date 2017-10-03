package sh

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestServerManualBackendPanic(t *testing.T) {
	_, stderr := run(t, "server_manual_backend")

	if strings.Contains(stderr, "panic") {
		t.Error("server panic")
	}
}

func run(t *testing.T, file string) (string, string) {
	var stdout, stderr bytes.Buffer

	defer func() {
		t.Log("STDOUT\n", stdout.String())
		t.Log("STDERR\n", stderr.String())
	}()

	timeout := time.After(time.Second * 10)
	done := make(chan struct{})
	go func() {
		select {
		case <-done:
			return
		case <-timeout:
			panic("timed out")
		}
	}()

	cmd := exec.Command("/bin/sh", file+".sh")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	close(done)

	if err != nil {
		t.Error(err)
	}
	return stdout.String(), stderr.String()
}
