package storage

import (
	"context"
	"github.com/minio/minio-go"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/tests"
	"io/ioutil"
	"testing"
)

func TestGenericS3Storage(t *testing.T) {
	tests.SetLogOutput(log, t)

	if !conf.Worker.Storage.S3[0].Valid() {
		t.Skipf("Skipping generic s3 e2e tests...")
	}

	store := storage.Storage{}
	store, err := store.WithConfig(conf.Worker.Storage)
	if err != nil {
		t.Fatal("error configuring storage:", err)
	}

	testBucket := "funnel-e2e-tests-" + tests.RandomString(6)

	sconf := conf.Worker.Storage
	client, err := minio.NewV2(sconf.GS3.Endpoint, sconf.GS3.Key, sconf.GS3.Secret, false)
	if err != nil {
		t.Fatal("error creating s3 client:", err)
	}

	err = client.MakeBucket(testBucket, "")
	if err != nil {
		t.Fatal("error creating test s3 bucket:", err)
	}

	defer func() {
		minioEmptyBucket(client, testBucket)
		client.RemoveBucket(testBucket)
	}()

	fPath := "testdata/test_in"
	inFileURL := "gs3://" + testBucket + "/" + fPath
	_, err = store.Put(context.Background(), inFileURL, fPath, tes.FileType_FILE)
	if err != nil {
		t.Fatal("error uploading test file:", err)
	}

	dPath := "testdata/test_dir"
	inDirURL := "gs3://" + testBucket + "/" + dPath
	_, err = store.Put(context.Background(), inDirURL, dPath, tes.FileType_DIRECTORY)
	if err != nil {
		t.Fatal("error uploading test directory:", err)
	}

	outFileURL := "gs3://" + testBucket + "/" + "test-output-file.txt"
	outDirURL := "gs3://" + testBucket + "/" + "test-output-directory"

	task := &tes.Task{
		Name: "s3 e2e",
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
					"cat $(find /opt/inputs -type f) > test-output-file.txt; mkdir test-output-directory; cp *.txt test-output-directory/",
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

	err = store.Get(context.Background(), outFileURL, "./test_tmp/test-s3-file.txt", tes.FileType_FILE)
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

	err = store.Get(context.Background(), outDirURL, "./test_tmp/test-s3-directory", tes.FileType_DIRECTORY)
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

	tests.SetLogOutput(log, t)
}

// minioEmptyBucket empties the S3 bucket
func minioEmptyBucket(client *minio.Client, bucket string) error {
	log.Info("Removing objects from S3 bucket : ", bucket)
	doneCh := make(chan struct{})
	defer close(doneCh)
	recursive := true
	for obj := range client.ListObjectsV2(bucket, "", recursive, doneCh) {
		client.RemoveObject(bucket, obj.Key)
	}
	log.Info("Emptied S3 bucket : ", bucket)
	return nil
}
