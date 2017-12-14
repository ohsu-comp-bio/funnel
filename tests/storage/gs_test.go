package storage

import (
	"context"
	"flag"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/tests"
	"golang.org/x/oauth2/google"
	gs "google.golang.org/api/storage/v1"
	"io/ioutil"
	"testing"
)

func TestGoogleStorage(t *testing.T) {
	tests.SetLogOutput(log, t)

	if !conf.GoogleStorage.Valid() {
		t.Skipf("Skipping google storage e2e tests...")
	}

	args := flag.Args()
	var projectID string
	if len(args) > 0 {
		projectID = args[0]
	}

	if projectID == "" {
		t.Fatal("Must provide projectID as an arg")
	}

	testBucket := "funnel-e2e-tests-" + tests.RandomString(6)

	client, err := newGsTest()
	if err != nil {
		t.Fatal(err)
	}
	err = client.createBucket(projectID, testBucket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		client.deleteBucket(testBucket)
	}()

	protocol := "gs://"

	store, err := storage.NewStorage(conf)
	if err != nil {
		t.Fatal("error configuring storage:", err)
	}

	fPath := "testdata/test_in"
	inFileURL := protocol + testBucket + "/" + fPath
	_, err = store.Put(context.Background(), inFileURL, fPath, tes.FileType_FILE)
	if err != nil {
		t.Fatal("error uploading test file:", err)
	}

	dPath := "testdata/test_dir"
	inDirURL := protocol + testBucket + "/" + dPath
	_, err = store.Put(context.Background(), inDirURL, dPath, tes.FileType_DIRECTORY)
	if err != nil {
		t.Fatal("error uploading test directory:", err)
	}

	outFileURL := protocol + testBucket + "/" + "test-output-file.txt"
	outDirURL := protocol + testBucket + "/" + "test-output-directory"

	task := &tes.Task{
		Name: "gs e2e",
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
		Executors: []*tes.Executor{
			{
				Image: "alpine:latest",
				Command: []string{
					"sh",
					"-c",
					"cat $(find /opt/inputs -type f | sort) > test-output-file.txt; mkdir test-output-directory; cp *.txt test-output-directory/",
				},
				Workdir: "/opt/workdir",
			},
		},
	}

	resp, err := fun.RPC.CreateTask(context.Background(), task)
	if err != nil {
		t.Fatal(err)
	}

	taskFinal := fun.Wait(resp.Id)

	if taskFinal.State != tes.State_COMPLETE {
		t.Fatal("Unexpected task failure")
	}

	expected := "file1 content\nfile2 content\nhello\n"

	err = store.Get(context.Background(), outFileURL, "./test_tmp/test-gs-file.txt", tes.FileType_FILE)
	if err != nil {
		t.Fatal("Failed to download file:", err)
	}

	b, err := ioutil.ReadFile("./test_tmp/test-gs-file.txt")
	if err != nil {
		t.Fatal("Failed to read downloaded file:", err)
	}
	actual := string(b)

	if actual != expected {
		t.Log("expected:", expected)
		t.Log("actual:  ", actual)
		t.Fatal("unexpected content")
	}

	err = store.Get(context.Background(), outDirURL, "./test_tmp/test-gs-directory", tes.FileType_DIRECTORY)
	if err != nil {
		t.Fatal("Failed to download directory:", err)
	}

	b, err = ioutil.ReadFile("./test_tmp/test-gs-directory/test-output-file.txt")
	if err != nil {
		t.Fatal("Failed to read file in downloaded directory", err)
	}
	actual = string(b)

	if actual != expected {
		t.Log("expected:", expected)
		t.Log("actual:  ", actual)
		t.Fatal("unexpected content")
	}
}

type gsTest struct {
	client *gs.Service
}

func newGsTest() (*gsTest, error) {
	defClient, err := google.DefaultClient(context.Background(), gs.CloudPlatformScope)
	if err != nil {
		return nil, err
	}
	client, err := gs.New(defClient)
	if err != nil {
		return nil, err
	}
	return &gsTest{client}, err
}

func (g *gsTest) createBucket(projectID, bucket string) error {
	req := g.client.Buckets.Insert(projectID, &gs.Bucket{
		Name: bucket,
	})
	_, err := req.Do()
	return err
}

func (g *gsTest) deleteBucket(bucket string) error {
	objects, err := g.client.Objects.List(bucket).Do()
	if err != nil {
		return err
	}
	for _, obj := range objects.Items {
		err = g.client.Objects.Delete(bucket, obj.Name).Do()
		if err != nil {
			return err
		}
	}
	return g.client.Buckets.Delete(bucket).Do()
}
