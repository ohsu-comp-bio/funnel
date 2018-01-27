package storage

import (
	"context"
	"github.com/ncw/swift"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/tests"
	"io/ioutil"
	"strings"
	"testing"
)

func TestSwiftStorage(t *testing.T) {
	tests.SetLogOutput(log, t)

	if !conf.Swift.Valid() {
		t.Skipf("Skipping swift e2e tests...")
	}

	testBucket := "funnel-e2e-tests-" + tests.RandomString(6)
	ctx := context.Background()

	client, err := newSwiftTest()
	if err != nil {
		t.Fatal("error creating swift client:", err)
	}
	err = client.createBucket(testBucket)
	if err != nil {
		t.Fatal("error creating test bucket:", err)
	}
	defer func() {
		client.deleteBucket(testBucket)
	}()

	protocol := "swift://"

	store, err := storage.NewStorage(conf)
	if err != nil {
		t.Fatal("error configuring storage:", err)
	}

	fPath := "testdata/test_in"
	inFileURL := protocol + testBucket + "/" + fPath
	_, err = store.Put(ctx, inFileURL, fPath, tes.FileType_FILE)
	if err != nil {
		t.Fatal("error uploading test file:", err)
	}

	dPath := "testdata/test_dir"
	inDirURL := protocol + testBucket + "/" + dPath
	_, err = store.Put(ctx, inDirURL, dPath, tes.FileType_DIRECTORY)
	if err != nil {
		t.Fatal("error uploading test directory:", err)
	}

	outFileURL := protocol + testBucket + "/" + "test-output-file.txt"
	outDirURL := protocol + testBucket + "/" + "test-output-directory"

	task := &tes.Task{
		Name: "storage e2e",
		Inputs: []*tes.Input{
			{
				Url:  inFileURL,
				Path: "/opt/inputs/test-file.txt",
				Type: tes.FileType_FILE,
			},
			{
				Url:  inDirURL,
				Path: "/opt/inputs/test-directory",
				Type: tes.FileType_DIRECTORY,
			},
		},
		Outputs: []*tes.Output{
			{
				Path: "/opt/workdir/test-output-file.txt",
				Url:  outFileURL,
				Type: tes.FileType_FILE,
			},
			{
				Path: "/opt/workdir/test-output-directory",
				Url:  outDirURL,
				Type: tes.FileType_DIRECTORY,
			},
		},
    Image: "alpine:latest",
    Command: []string{
      "sh",
      "-c",
      "cat $(find /opt/inputs -type f | sort) > test-output-file.txt; mkdir test-output-directory; cp *.txt test-output-directory/",
    },
    Workdir: "/opt/workdir",
	}

	resp, err := fun.RPC.CreateTask(ctx, task)
	if err != nil {
		t.Fatal(err)
	}

	taskFinal := fun.Wait(resp.Id)

	if taskFinal.State != tes.State_COMPLETE {
		t.Fatal("Unexpected task failure")
	}

	expected := "file1 content\nfile2 content\nhello\n"

	err = store.Get(ctx, outFileURL, "./test_tmp/test-s3-file.txt", tes.FileType_FILE)
	if err != nil {
		t.Fatal("Failed to download file:", err)
	}

	b, err := ioutil.ReadFile("./test_tmp/test-s3-file.txt")
	if err != nil {
		t.Fatal("Failed to read downloaded file:", err)
	}
	actual := string(b)

	if actual != expected {
		t.Log("expected:", expected)
		t.Log("actual:  ", actual)
		t.Fatal("unexpected content")
	}

	err = store.Get(ctx, outDirURL, "./test_tmp/test-s3-directory", tes.FileType_DIRECTORY)
	if err != nil {
		t.Fatal("Failed to download directory:", err)
	}

	b, err = ioutil.ReadFile("./test_tmp/test-s3-directory/test-output-file.txt")
	if err != nil {
		t.Fatal("Failed to read file in downloaded directory", err)
	}
	actual = string(b)

	if actual != expected {
		t.Log("expected:", expected)
		t.Log("actual:  ", actual)
		t.Fatal("unexpected content")
	}

	// should succeed with warning when there is an input or output directory that
	// does not exist
	task = &tes.Task{
		Name: "storage e2e",
		Inputs: []*tes.Input{
			{
				Url:  protocol + testBucket + "/this/path/does/not/exist",
				Path: "/opt/inputs/test-directory",
				Type: tes.FileType_DIRECTORY,
			},
		},
		Outputs: []*tes.Output{
			{
				Path: "/opt/workdir/this/path/does/not/exist/test-output-directory",
				Url:  outDirURL,
				Type: tes.FileType_DIRECTORY,
			},
		},
    Image: "alpine:latest",
    Command: []string{
      "sleep", "1",
    },
	}

	resp, err = fun.RPC.CreateTask(ctx, task)
	if err != nil {
		t.Fatal(err)
	}

	taskFinal = fun.Wait(resp.Id)
	if taskFinal.State != tes.State_COMPLETE {
		t.Fatal("Expected task failure")
	}
	found := false
	for _, log := range taskFinal.Logs[0].SystemLogs {
		if strings.Contains(log, "level='warning'") {
			found = true
		}
	}
	if !found {
		t.Fatal("Expected warning in system logs")
	}
}

type swiftTest struct {
	client *swift.Connection
}

func newSwiftTest() (*swiftTest, error) {
	conn := &swift.Connection{}
	err := conn.ApplyEnvironment()
	if err != nil {
		return nil, err
	}
	err = conn.Authenticate()
	if err != nil {
		return nil, err
	}
	return &swiftTest{conn}, nil
}

func (sw *swiftTest) createBucket(bucket string) error {
	return sw.client.ContainerCreate(bucket, nil)
}

func (sw *swiftTest) deleteBucket(bucket string) error {
	objs, err := sw.client.ObjectsAll(bucket, nil)
	if err != nil {
		return err
	}
	for _, obj := range objs {
		err = sw.client.ObjectDelete(bucket, obj.Name)
		if err != nil {
			return err
		}
	}
	return sw.client.ContainerDelete(bucket)
}
