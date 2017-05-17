package storage

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

func TestLocalSupports(t *testing.T) {
	l := LocalBackend{}
	if !l.Supports("file:///path/to/foo.txt", "/host/path", tes.FileType_FILE) {
		t.Fatal("Expected file:// URL to be supported")
	}

	if !l.Supports("/path/to/foo.txt", "/host/path", tes.FileType_FILE) {
		t.Fatal("Expected normal file path to be supported")
	}
}

// Tests Get on a "file://" URL, e.g. "file:///path/to/foo.txt"
func TestLocalGet(t *testing.T) {
	ctx := context.Background()
	tmp, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	logger.Debug("TEMP DIR", tmp)
	l := LocalBackend{allowedDirs: []string{tmp}}
	ip := path.Join(tmp, "input.txt")
	cp := path.Join(tmp, "container.txt")
	ioutil.WriteFile(ip, []byte("foo"), os.ModePerm)
	gerr := l.Get(ctx, "file://"+ip, cp, tes.FileType_FILE)
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
	logger.Debug("TEMP DIR", tmp)
	l := LocalBackend{allowedDirs: []string{tmp}}
	ip := path.Join(tmp, "input.txt")
	cp := path.Join(tmp, "container.txt")
	ioutil.WriteFile(ip, []byte("foo"), os.ModePerm)
	gerr := l.Get(ctx, ip, cp, tes.FileType_FILE)
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
	logger.Debug("TEMP DIR", tmp)
	l := LocalBackend{allowedDirs: []string{tmp}}

	// Write the test files
	cp := path.Join(tmp, "container.txt")
	op := path.Join(tmp, "output.txt")
	ioutil.WriteFile(cp, []byte("foo"), os.ModePerm)

	gerr := l.Put(ctx, "file://"+op, cp, tes.FileType_FILE)
	if gerr != nil {
		t.Fatal(gerr)
	}

	// Check the resulting content
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
	logger.Debug("TEMP DIR", tmp)
	l := LocalBackend{allowedDirs: []string{tmp}}

	// Write the test files
	cp := path.Join(tmp, "container.txt")
	op := path.Join(tmp, "output.txt")
	ioutil.WriteFile(cp, []byte("foo"), os.ModePerm)

	gerr := l.Put(ctx, op, cp, tes.FileType_FILE)
	if gerr != nil {
		t.Fatal(gerr)
	}

	// Check the resulting content
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
	ctx := context.Background()
	tmp, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	logger.Debug("TEMP DIR", tmp)
	tmpOut, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	logger.Debug("TEMP OUT DIR", tmpOut)
	l := LocalBackend{allowedDirs: []string{tmp, tmp_out}}

	// Write the test files
	cp := path.Join(tmp, "output.txt")
	op := path.Join(tmpOut, "output.txt")
	ioutil.WriteFile(cp, []byte("foo"), os.ModePerm)

	pferr := l.Put(ctx, op, cp, tes.FileType_FILE)
	if pferr != nil {
		t.Fatal(pferr)
	}

	// since the file in this dir were already 'Put' in the previous step
	// nothing should happen
	pderr := l.Put(ctx, tmpOut, tmp, tes.FileType_DIRECTORY)
	if pderr != nil {
		t.Fatal(pderr)
	}

	// Check the resulting content
	b, rerr := ioutil.ReadFile(op)
	if rerr != nil {
		t.Fatal(rerr)
	}
	if string(b) != "foo" {
		t.Fatal("Unexpected content")
	}
}
