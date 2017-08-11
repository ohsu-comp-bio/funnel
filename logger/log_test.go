package logger

import (
	"bytes"
	"context"
	"errors"
	"testing"
)

func TestLog(t *testing.T) {
	l := New("foons", "basearg", 1)
	c := DefaultConfig()
	c.JSONFormat.DisableTimestamp = true
	l.Configure(c)

	var b bytes.Buffer
	l.SetOutput(&b)
	l.Info("test")

	expect := `{"basearg":1,"level":"info","msg":"test","ns":"foons"}` + "\n"
	if b.String() != expect {
		t.Fatal("unexpected log:", b.String())
	}
}

func TestContextLog(t *testing.T) {
	l := New("foons", "basearg", 1)
	c := DefaultConfig()
	c.JSONFormat.DisableTimestamp = true
	l.Configure(c)

	var b bytes.Buffer
	l.SetOutput(&b)

	ctx := context.WithValue(context.Background(), TaskIDKey, "task-1")
	l.Info("test", ctx)

	expect := `{"basearg":1,"level":"info","msg":"test","ns":"foons","taskID":"task-1"}` + "\n"
	if b.String() != expect {
		t.Fatal("unexpected log:", b.String())
	}
}

func TestErrorFieldLog(t *testing.T) {
	l := New("foons", "basearg", 1)
	c := DefaultConfig()
	c.JSONFormat.DisableTimestamp = true
	l.Configure(c)

	var b bytes.Buffer
	l.SetOutput(&b)

	err := errors.New("fooerr")
	l.Info("test", err)

	expect := `{"basearg":1,"error":"fooerr","level":"info","msg":"test","ns":"foons"}` + "\n"
	if b.String() != expect {
		t.Fatal("unexpected log:", b.String())
	}
}
