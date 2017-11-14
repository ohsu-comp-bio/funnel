package storage

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/tests/e2e"
	"io/ioutil"
	"testing"
)

func TestS3StorageTask(t *testing.T) {
	e2e.SetLogOutput(log, t)
	if conf.Worker.Storage.S3.Valid() {
		t.Skipf("Skipping s3 e2e tests...")
	}

	task := &tes.Task{
		Name: "s3 e2e",
		Inputs: []*tes.Input{
			{
				Url:  "s3://strucka-dev/test-file.txt",
				Path: "/opt/inputs/test-file.txt",
				Type: tes.FileType_FILE,
			},
			{
				Url:  "s3://strucka-dev/test-directory",
				Path: "/opt/inputs/test-directory",
				Type: tes.FileType_DIRECTORY,
			},
		},
		Outputs: []*tes.Output{
			{
				Path: "/opt/workdir/test-output-file.txt",
				Url:  "s3://strucka-dev/test_tmp/test-output-file.txt",
				Type: tes.FileType_FILE,
			},
			{
				Path: "/opt/workdir/test-output-directory",
				Url:  "s3://strucka-dev/test_tmp/test-output-directory",
				Type: tes.FileType_DIRECTORY,
			},
		},
		Executors: []*tes.Executor{
			{
				Image: "alpine:latest",
				Command: []string{
					"sh",
					"-c",
					"echo $(find /opt/inputs -type f) > test-output-file.txt; mkdir test-output-directory; cp *.txt test-output-directory/",
				},
				Workdir: "/opt/workdir",
			},
		},
	}

	ctx := context.Background()

	resp, err := fun.RPC.CreateTask(ctx, task)
	if err != nil {
		t.Fatal(err)
	}

	taskFinal := fun.Wait(resp.Id)

	if taskFinal.State != tes.State_COMPLETE {
		t.Fatal("Unexpected task failure")
	}

	expected := "/opt/inputs/test-directory/bar.txt /opt/inputs/test-directory/foo.txt /opt/inputs/test-file.txt\n"

	s := storage.Storage{}
	s, err = s.WithConfig(conf.Worker.Storage)
	if err != nil {
		t.Fatal("Error configuring storage", err)
	}

	err = s.Get(ctx, "s3://strucka-dev/test_tmp/test-output-file.txt", "./test_tmp/test-s3-file.txt", tes.FileType_FILE)
	if err != nil {
		t.Fatal("Failed get", err)
	}

	b, err := ioutil.ReadFile("./test_tmp/test-s3-file.txt")
	if err != nil {
		t.Fatal("Failed read", err)
	}

	actual := string(b)

	if actual != expected {
		t.Log("expected:", expected)
		t.Log("actual:  ", actual)
		t.Fatal("unexpected content")
	}

	err = s.Get(ctx, "s3://strucka-dev/test_tmp/test-output-directory", "./test_tmp/test-s3-directory", tes.FileType_DIRECTORY)
	if err != nil {
		t.Fatal("Failed get", err)
	}

	b, err = ioutil.ReadFile("./test_tmp/test-s3-directory/test-output-file.txt")
	if err != nil {
		t.Fatal("Failed read", err)
	}

	actual = string(b)

	if actual != expected {
		t.Log("expected:", expected)
		t.Log("actual:  ", actual)
		t.Fatal("unexpected content")
	}
}
