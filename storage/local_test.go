package storage

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestLocalSupports(t *testing.T) {
	l := LocalBackend{allowedDirs: []string{"/path"}}

	err := l.SupportsGet("file:///path/to/foo.txt", File)
	if err != nil {
		t.Fatal("Expected file:// URL to be supported", err)
	}

	err = l.SupportsGet("/path/to/foo.txt", File)
	if err != nil {
		t.Fatal("Expected normal file path to be supported", err)
	}
}

// Tests Get on a "file://" URL, e.g. "file:///path/to/foo.txt"
func TestLocalGet(t *testing.T) {
	ctx := context.Background()
	tmp, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	l := Storage{}.WithBackend(&LocalBackend{allowedDirs: []string{tmp}})

	// File test
	ip := path.Join(tmp, "input.txt")
	cp := path.Join(tmp, "container.txt")
	ioutil.WriteFile(ip, []byte("foo"), os.ModePerm)

	gerr := l.Get(ctx, "file://"+ip, cp, File)
	if gerr != nil {
		t.Fatal(gerr)
	}

	b, rerr := ioutil.ReadFile(cp)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(b) != "foo" {
		t.Fatal("Unexpected content")
	}

	// Directory test
	id := path.Join(tmp, "subdir")
	os.MkdirAll(id, os.ModePerm)
	idf := path.Join(id, "other.txt")
	cd := path.Join(tmp, "localized_dir")
	ioutil.WriteFile(idf, []byte("bar"), os.ModePerm)

	gerr = l.Get(ctx, "file://"+id, cd, Directory)
	if gerr != nil {
		t.Fatal(gerr)
	}

	b, rerr = ioutil.ReadFile(path.Join(cd, "other.txt"))
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(b) != "bar" {
		t.Fatal("Unexpected content")
	}
}

// Tests Get on a URL that is a path, e.g. "/path/to/foo.txt"
func TestLocalGetPath(t *testing.T) {
	ctx := context.Background()
	tmp, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	l := Storage{}.WithBackend(&LocalBackend{allowedDirs: []string{tmp}})

	ip := path.Join(tmp, "input.txt")
	cp := path.Join(tmp, "container.txt")
	ioutil.WriteFile(ip, []byte("foo"), os.ModePerm)

	gerr := l.Get(ctx, ip, cp, File)
	if gerr != nil {
		t.Fatal(gerr)
	}

	b, rerr := ioutil.ReadFile(cp)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(b) != "foo" {
		t.Fatal("Unexpected content")
	}
}

// Tests Put on a "file://" URL, e.g. "file:///path/to/foo.txt"
func TestLocalPut(t *testing.T) {
	ctx := context.Background()
	tmp, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	l := Storage{}.WithBackend(&LocalBackend{allowedDirs: []string{tmp}})

	// File test
	cp := path.Join(tmp, "container.txt")
	op := path.Join(tmp, "output.txt")
	ioutil.WriteFile(cp, []byte("foo"), os.ModePerm)

	_, gerr := l.Put(ctx, "file://"+op, cp, File)
	if gerr != nil {
		t.Fatal(gerr)
	}

	b, rerr := ioutil.ReadFile(op)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(b) != "foo" {
		t.Fatal("Unexpected content")
	}

	// Directory test
	cd := path.Join(tmp, "subdir")
	os.MkdirAll(cd, os.ModePerm)
	cdf := path.Join(cd, "other.txt")
	od := path.Join(tmp, "subout")
	ioutil.WriteFile(cdf, []byte("bar"), os.ModePerm)
	_, gerr = l.Put(ctx, "file://"+od, cd, Directory)
	if gerr != nil {
		t.Fatal(gerr)
	}

	b, rerr = ioutil.ReadFile(path.Join(od, "other.txt"))
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(b) != "bar" {
		t.Fatal("Unexpected content")
	}
}

// Tests Put on a URL that is a path, e.g. "/path/to/foo.txt"
func TestLocalPutPath(t *testing.T) {
	ctx := context.Background()
	tmp, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	l := Storage{}.WithBackend(&LocalBackend{allowedDirs: []string{tmp}})

	cp := path.Join(tmp, "container.txt")
	op := path.Join(tmp, "output.txt")
	ioutil.WriteFile(cp, []byte("foo"), os.ModePerm)

	_, gerr := l.Put(ctx, op, cp, File)
	if gerr != nil {
		t.Fatal(gerr)
	}

	b, rerr := ioutil.ReadFile(op)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(b) != "foo" {
		t.Fatal("Unexpected content")
	}
}

// Tests Put when source and dest reference the same file (inode)
// Since the LocalBackend hard-links files when possible we need to protect
// against the case where the same path is 'Put' twice
func TestSameFile(t *testing.T) {
	tmp, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	tmpOut, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}

	// Write the test files
	cp := path.Join(tmp, "output.txt")
	cp2 := path.Join(tmp, "output2.txt")
	op := path.Join(tmpOut, "output.txt")
	ioutil.WriteFile(cp, []byte("foo"), os.ModePerm)
	ioutil.WriteFile(cp2, []byte("bar"), os.ModePerm)

	err = linkFile(cp, op)
	if err != nil {
		t.Fatal(err)
	}

	// since the file in this dir were already 'Put' in the previous step
	// nothing should happen
	err = copyFile(cp, op)
	if err != nil {
		t.Fatal(err)
	}

	// Check the resulting content
	b, err := ioutil.ReadFile(op)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "foo" {
		t.Fatal("Unexpected content")
	}

	// same file output url; new src contents
	err = copyFile(cp2, op)
	if err != nil {
		t.Fatal(err)
	}

	// Check the resulting content
	b, err = ioutil.ReadFile(op)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "bar" {
		t.Fatal("Unexpected content")
	}
}

func TestGetPath(t *testing.T) {
	x := "file:///foo/bar/file with spaces.txt"
	e := getPath(x)
	if e != "/foo/bar/file with spaces.txt" {
		t.Fatal("Unexpected URL encoding")
	}

	x = "/foo/bar/escaped%20with%20.txt"
	e = getPath(x)
	if e != "/foo/bar/escaped%20with%20.txt" {
		t.Fatal("Unexpected URL encoding")
	}
}
