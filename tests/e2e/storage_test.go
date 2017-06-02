package e2e

import (
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"os"
	"syscall"
	"testing"
)

// Test that a file can be passed as an input and output.
func TestFileMount(t *testing.T) {
	id := fun.Run(`
    --cmd "sh -c 'cat $in > $out'"
    -i in=./testdata/test_in
    -o out={{ .storage }}/test_out
  `)
	task := fun.Wait(id)
	c := fun.ReadFile("test_out")
	if c != "hello\n" {
		log.Error("TASK", task)
		log.Error("CONTENT", c)
		t.Fatal("Unexpected output file")
	}
}

// Test that the local storage system hard links input files.
// TODO this test is unix specific because it uses syscall?
func TestLocalFilesystemHardLinkInput(t *testing.T) {
	fun.WriteFile("test_hard_link_input", "content")
	id := fun.Run(`
    --cmd "echo foo"
    -i in={{ .storage }}/test_hard_link_input
  `)
	task := fun.Wait(id)
	if task.State != tes.State_COMPLETE {
		t.Fatal("unexpected task failure")
	}
	name := fun.StorageDir + "/test_hard_link_input"
	fi, sterr := os.Lstat(name)
	if sterr != nil {
		panic(sterr)
	}
	s, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		panic("can't retrieve Stat_t")
	}
	if uint16(s.Nlink) != uint16(2) {
		t.Fatal("expected to links")
	}
}

// Test using a symlink as an input file.
func TestSymlinkInput(t *testing.T) {
	id := fun.Run(`
    --cmd "sh -c 'cat $in > $out'"
    -i in=./testdata/test_in_symlink
    -o out={{ .storage }}/test_out
  `)
	task := fun.Wait(id)
	if task.State != tes.State_COMPLETE {
		t.Fatal("Expected success on symlink input")
	}
}

// Test using a broken symlink as an input file.
func TestBrokenSymlinkInput(t *testing.T) {
	id := fun.Run(`
    --cmd "sh -c 'cat $in > $out'"
    -i in=./testdata/test_broken_symlink
    -o out={{ .storage }}/test_out
  `)
	task := fun.Wait(id)
	if task.State != tes.State_SYSTEM_ERROR {
		t.Fatal("Expected error on broken symlink input")
	}
}

/*
  Test the case where a container creates a symlink in an output path.
  From the view of the host system where Funnel is running, this creates
  a broken link, because the source of the symlink is a path relative
  to the container filesystem.

  Funnel can fix some of these cases using volume definitions, which
  is being tested here.
*/
func TestSymlinkOutput(t *testing.T) {
	id := fun.Run(`
    --cmd "sh -c 'echo foo > $dir/foo && ln -s $dir/foo $dir/sym && ln -s $dir/foo $sym'"
    -o sym={{ .storage }}/out-sym
    -O dir={{ .storage }}/out-dir
  `)
	task := fun.Wait(id)

	if task.State != tes.State_COMPLETE {
		t.Fatal("expected success on symlink output")
	}

	if fun.ReadFile("out-dir/foo") != "foo\n" {
		t.Fatal("unexpected out-dir/foo content")
	}

	if fun.ReadFile("out-sym") != "foo\n" {
		t.Fatal("unexpected out-sym content")
	}

	if fun.ReadFile("out-dir/sym") != "foo\n" {
		t.Fatal("unexpected out-dir/sym content")
	}
}

func TestOverwriteOutput(t *testing.T) {
	id := fun.Run(`
    --cmd "sh -c 'echo foo > $out; chmod go+w $out'"
    -o out={{ .storage }}/test_out
  `)
	task := fun.Wait(id)

	if task.State != tes.State_COMPLETE {
		t.Fatal("expected success")
	}

	// this time around since the output file exists copyFile will be called
	// in storage.Put
	id = fun.Run(`
    --cmd "sh -c 'echo foo > $out; chmod go+w $out'"
    -o out={{ .storage }}/test_out
  `)
	task = fun.Wait(id)

	if task.State != tes.State_COMPLETE {
		t.Fatal("expected success")
	}
}
