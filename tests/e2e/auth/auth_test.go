package e2e

import (
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"github.com/ohsu-comp-bio/funnel/util"
	"os"
	"strings"
	"testing"
)

var log = logger.Sub("e2e-auth")

var extask = `{
  "executors": [
    {
      "image_name": "alpine",
      "cmd": ["echo", "hello world"]
    }
  ]
}`

func TestBasicAuthFail(t *testing.T) {
	conf := e2e.DefaultConfig()
	conf.Server.Password = "abc123"
	fun := e2e.NewFunnel(conf)
	fun.WithLocalBackend()
	fun.StartServer()

	var err error
	_, err = fun.HTTP.GetTask("1", "MINIMAL")
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		log.Debug("ERR", err)
		t.Fatal("expected error")
	}

	_, err = fun.HTTP.ListTasks("MINIMAL")
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		log.Debug("ERR", err)
		t.Fatal("expected error")
	}

	_, err = fun.HTTP.CreateTask([]byte(extask))
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		log.Debug("ERR", err)
		t.Fatal("expected error")
	}

	_, err = fun.HTTP.CancelTask("1")
	if err == nil || !strings.Contains(err.Error(), "STATUS CODE - 403") {
		log.Debug("ERR", err)
		t.Fatal("expected error")
	}

	_, err = fun.RunE(`--cmd 'echo hello'`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestBasicAuthed(t *testing.T) {
	os.Setenv("FUNNEL_SERVER_PASSWORD", "abc123")
	defer os.Unsetenv("FUNNEL_SERVER_PASSWORD")
	conf := e2e.DefaultConfig()
	conf.Server.Password = "abc123"
	fun := e2e.NewFunnel(conf)
	fun.WithLocalBackend()
	fun.StartServer()
	fun.AddRPCClient(util.PerRPCPassword("abc123"))

	var err error
	log.Debug("CLI", fun.HTTP)

	// Run a task to completion
	id2 := fun.Run(`--cmd 'echo hello'`)
	t2 := fun.Wait(id2)
	if t2.State != tes.State_COMPLETE {
		t.Fatal("expected task to complete")
	}

	_, err = fun.HTTP.GetTask(id2, "MINIMAL")
	if err != nil {
		t.Fatal("unexpected error")
	}

	_, err = fun.HTTP.ListTasks("MINIMAL")
	if err != nil {
		t.Fatal("unexpected error")
	}

	resp, err := fun.HTTP.CreateTask([]byte(extask))
	if err != nil {
		t.Fatal("unexpected error")
	}

	_, err = fun.HTTP.CancelTask(resp.Id)
	if err != nil {
		t.Fatal("unexpected error")
	}
}
