package storage

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestLocalSupports(t *testing.T) {
	l := Local{allowedDirs: []string{"/path"}}

	err := l.UnsupportedOperations("file:///path/to/foo.txt").Get
	if err != nil {
		t.Fatal("Expected file:// URL to be supported", err)
	}

	err = l.UnsupportedOperations("/path/to/foo.txt").Get
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
	l := &Local{allowedDirs: []string{tmp}}

	// File test
	ip := path.Join(tmp, "input.txt")
	cp := path.Join(tmp, "container.txt")
	ioutil.WriteFile(ip, []byte("foo"), os.ModePerm)

	_, gerr := l.Get(ctx, "file://"+ip, cp)
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

// Tests Get on a URL that is a path, e.g. "/path/to/foo.txt"
func TestLocalGetPath(t *testing.T) {
	ctx := context.Background()
	tmp, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	l := &Local{allowedDirs: []string{tmp}}

	ip := path.Join(tmp, "input.txt")
	cp := path.Join(tmp, "container.txt")
	ioutil.WriteFile(ip, []byte("foo"), os.ModePerm)

	_, gerr := l.Get(ctx, ip, cp)
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
	l := &Local{allowedDirs: []string{tmp}}

	// File test
	cp := path.Join(tmp, "container.txt")
	op := path.Join(tmp, "output.txt")
	ioutil.WriteFile(cp, []byte("foo"), os.ModePerm)

	_, gerr := l.Put(ctx, "file://"+op, cp)
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

// Tests Put on a URL that is a path, e.g. "/path/to/foo.txt"
func TestLocalPutPath(t *testing.T) {
	ctx := context.Background()
	tmp, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	l := &Local{allowedDirs: []string{tmp}}

	cp := path.Join(tmp, "container.txt")
	op := path.Join(tmp, "output.txt")
	ioutil.WriteFile(cp, []byte("foo"), os.ModePerm)

	_, gerr := l.Put(ctx, op, cp)
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
// Since Local storage hard-links files when possible we need to protect
// against the case where the same path is 'Put' twice
func TestSameFile(t *testing.T) {
	ctx := context.Background()
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

	err = linkFile(ctx, cp, op)
	if err != nil {
		t.Fatal(err)
	}

	// since the file in this dir were already 'Put' in the previous step
	// nothing should happen
	err = copyFile(ctx, cp, op)
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
	err = copyFile(ctx, cp2, op)
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
