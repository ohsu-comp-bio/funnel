package task

import (
	"io"
	"os"
	"testing"
)

func TestGet(t *testing.T) {
	cmd, h := newCommandHooks()

	h.Get = func(server string, ids []string, view string, w io.Writer) error {
		if view != "MINIMAL" {
			t.Errorf("expected MINIMAL view, got '%s'", view)
		}
		if ids[0] != "1" || ids[1] != "2" || len(ids) != 2 {
			t.Errorf("unexpected ids: %#v", ids)
		}
		return nil
	}

	cmd.SetArgs([]string{"get", "--view", "MINIMAL", "1", "2"})
	cmd.Execute()
}

// "get" command should have default view of FULL
func TestGetDefaultView(t *testing.T) {
	cmd, h := newCommandHooks()

	h.Get = func(server string, ids []string, view string, w io.Writer) error {
		if view != "full" {
			t.Errorf("expected default 'full' view, got '%s'", view)
		}
		return nil
	}

	cmd.SetArgs([]string{"get", "1", "2"})
	cmd.Execute()
}

func TestList(t *testing.T) {
	cmd, h := newCommandHooks()

	h.List = func(server, view, page, state string, tags []string, size int32, all bool, w io.Writer) error {
		if view != "FULL" {
			t.Errorf("expected FULL view, got '%s'", view)
		}
		return nil
	}

	cmd.SetArgs([]string{"list", "--view", "FULL"})
	cmd.Execute()
}

// Test that the server URL defaults to localhost:8000
func TestServerDefault(t *testing.T) {
	cmd, h := newCommandHooks()

	h.Create = func(server string, messages []string, r io.Reader, w io.Writer) error {
		if server != "http://localhost:8000" {
			t.Errorf("expected localhost default, got '%s'", server)
		}
		return nil
	}
	h.List = func(server, view, page, state string, tags []string, size int32, all bool, w io.Writer) error {
		if server != "http://localhost:8000" {
			t.Errorf("expected localhost default, got '%s'", server)
		}
		return nil
	}
	h.Get = func(server string, ids []string, view string, w io.Writer) error {
		if server != "http://localhost:8000" {
			t.Errorf("expected localhost default, got '%s'", server)
		}
		return nil
	}
	h.Cancel = func(server string, ids []string, w io.Writer) error {
		if server != "http://localhost:8000" {
			t.Errorf("expected localhost default, got '%s'", server)
		}
		return nil
	}
	h.Wait = func(server string, ids []string) error {
		if server != "http://localhost:8000" {
			t.Errorf("expected localhost default, got '%s'", server)
		}
		return nil
	}

	cmd.SetArgs([]string{"create", "foo.json"})
	cmd.Execute()

	cmd.SetArgs([]string{"list"})
	cmd.Execute()

	cmd.SetArgs([]string{"get", "1"})
	cmd.Execute()

	cmd.SetArgs([]string{"cancel", "1"})
	cmd.Execute()

	cmd.SetArgs([]string{"wait", "1"})
	cmd.Execute()
}

// Test that the server URL may be set via a FUNNEL_SERVER environment
// variable.
func TestServerEnv(t *testing.T) {
	os.Setenv("FUNNEL_SERVER", "foobar")

	cmd, h := newCommandHooks()

	h.Create = func(server string, messages []string, r io.Reader, w io.Writer) error {
		if server != "foobar" {
			t.Error("expected foobar")
		}
		return nil
	}
	h.List = func(server, view, page, state string, tags []string, size int32, all bool, w io.Writer) error {
		if server != "foobar" {
			t.Error("expected foobar")
		}
		return nil
	}
	h.Get = func(server string, ids []string, view string, w io.Writer) error {
		if server != "foobar" {
			t.Error("expected foobar")
		}
		return nil
	}
	h.Cancel = func(server string, ids []string, w io.Writer) error {
		if server != "foobar" {
			t.Error("expected foobar")
		}
		return nil
	}
	h.Wait = func(server string, ids []string) error {
		if server != "foobar" {
			t.Error("expected foobar")
		}
		return nil
	}

	cmd.SetArgs([]string{"create", "foo.json"})
	cmd.Execute()

	cmd.SetArgs([]string{"list"})
	cmd.Execute()

	cmd.SetArgs([]string{"get", "1"})
	cmd.Execute()

	cmd.SetArgs([]string{"cancel", "1"})
	cmd.Execute()

	cmd.SetArgs([]string{"wait", "1"})
	cmd.Execute()
}

// Test that the server flag overrides the FUNNEL_SERVER env var
func TestServerFlagOverride(t *testing.T) {
	os.Setenv("FUNNEL_SERVER", "foobar")
	srv := "flagval"

	cmd, h := newCommandHooks()

	h.Create = func(server string, messages []string, r io.Reader, w io.Writer) error {
		if server != "flagval" {
			t.Error("expected flagval")
		}
		return nil
	}
	h.List = func(server, view, page, state string, tags []string, size int32, all bool, w io.Writer) error {
		if server != "flagval" {
			t.Error("expected flagval")
		}
		return nil
	}
	h.Get = func(server string, ids []string, view string, w io.Writer) error {
		if server != "flagval" {
			t.Error("expected flagval")
		}
		return nil
	}
	h.Cancel = func(server string, ids []string, w io.Writer) error {
		if server != "flagval" {
			t.Error("expected flagval")
		}
		return nil
	}
	h.Wait = func(server string, ids []string) error {
		if server != "flagval" {
			t.Error("expected flagval")
		}
		return nil
	}

	cmd.SetArgs([]string{"create", "-S", srv, "foo.json"})
	cmd.Execute()

	cmd.SetArgs([]string{"list", "-S", srv})
	cmd.Execute()

	cmd.SetArgs([]string{"get", "-S", srv, "1"})
	cmd.Execute()

	cmd.SetArgs([]string{"cancel", "-S", srv, "1"})
	cmd.Execute()

	cmd.SetArgs([]string{"wait", "-S", srv, "1"})
	cmd.Execute()
}
