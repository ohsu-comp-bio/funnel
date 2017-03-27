package worker

import (
	"io/ioutil"
	"os"
	"path"
	pbe "tes/ga4gh"
	"testing"
)

func TestResolveLinks(t *testing.T) {
	// Setup
	f, _ := ioutil.TempDir("", "funnel-test-resolve-links-")
	m := NewFileMapper(f)
	r := jobRunner{
		mapper: m,
	}

	// Create file in container which
	ioutil.WriteFile(path.Join(f, "test-file"), []byte("foo\n"), 0777)

	// Create broken symlink. Simulates broken symlink from container
	// filesystem.
	os.Symlink("/mnt/foo/test-file", path.Join(f, "test-sym"))

	m.Volumes = append(m.Volumes, Volume{
		HostPath:      f,
		ContainerPath: "/mnt/foo",
	})

	m.Outputs = append(m.Outputs, &pbe.TaskParameter{
		Path: path.Join(f, "test-sym"),
	})
	// Include normal file in order to test that they are ignored
	m.Outputs = append(m.Outputs, &pbe.TaskParameter{
		Path: path.Join(f, "test-file"),
	})
	r.resolveLinks()

	c, e := ioutil.ReadFile(m.Outputs[0].Path)

	if e != nil {
		log.Error("Error reading file", e)
		t.Error("Error reading file")
		return
	}

	if string(c) != "foo\n" {
		t.Error("Error: unexpected content")
	}
}
