package storage

import (
	"context"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/tests"
)

func TestMultiS3Storage(t *testing.T) {
	tests.SetLogOutput(log, t)

	if len(conf.GenericS3) == 2 {
		if !conf.GenericS3[0].Valid() && !conf.GenericS3[1].Valid() {
			t.Skipf("Skipping generic s3 e2e tests...")
		}
	} else {
		t.Skipf("Skipping generic s3 e2e tests...")
	}

	testBucket := "funnel-e2e-tests-" + tests.RandomString(6)

	// Generic S3 backend setup
	gconf1 := conf.GenericS3[0]
	gclient1, err := newMinioTest(gconf1)
	if err != nil {
		t.Fatal("error creating minio client:", err)
	}
	err = gclient1.createBucket(testBucket)
	if err != nil {
		t.Fatal("error creating test bucket:", err)
	}
	defer func() {
		gclient1.deleteBucket(testBucket)
	}()

	gconf2 := conf.GenericS3[1]
	gclient2, err := newMinioTest(gconf2)
	if err != nil {
		t.Fatal("error creating minio client:", err)
	}
	err = gclient2.createBucket(testBucket)
	if err != nil {
		t.Fatal("error creating test bucket:", err)
	}
	defer func() {
		gclient2.deleteBucket(testBucket)
	}()

	// Stage input files
	protocol := "s3://"
	fPath := "testdata/test_in"

	g1FileURL := protocol + gconf1.Endpoint + "/" + testBucket + "/" + fPath + tests.RandomString(6)
	err = gclient1.fcli.PutFile(context.Background(), g1FileURL, fPath)
	if err != nil {
		t.Fatal("error uploading test file:", err)
	}

	g2FileURL := protocol + gconf2.Endpoint + "/" + testBucket + "/" + fPath + tests.RandomString(6)
	err = gclient2.fcli.PutFile(context.Background(), g2FileURL, fPath)
	if err != nil {
		t.Fatal("error uploading test file:", err)
	}

	// Expect the following task to complete since s3 urls contain endpoints
	outFileURL := protocol + gconf1.Endpoint + "/" + testBucket + "/" + "test-output-file.txt"
	task := &tes.Task{
		Name: "s3 e2e",
		Inputs: []*tes.Input{
			{
				Url:  g1FileURL,
				Path: "/opt/inputs/test-file1.txt",
				Type: tes.FileType_FILE,
			},
			{
				Url:  g2FileURL,
				Path: "/opt/inputs/test-file2.txt",
				Type: tes.FileType_FILE,
			},
		},
		Outputs: []*tes.Output{
			{
				Path: "/opt/workdir/test-output-file.txt",
				Url:  outFileURL,
				Type: tes.FileType_FILE,
			},
		},
		Executors: []*tes.Executor{
			{
				Image: "alpine:latest",
				Command: []string{
					"sh",
					"-c",
					"cat $(find /opt/inputs -type f | sort) > test-output-file.txt",
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

	expected := "hello\nhello\n"

	err = gclient1.fcli.Get(context.Background(), outFileURL, "./test_tmp/test-s3-file.txt", tes.FileType_FILE)
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

	// Expect the following task to fail due to s3 provider ambiguity
	g1FileURL = strings.Replace(g1FileURL, gconf1.Endpoint, "", -1)
	g2FileURL = strings.Replace(g2FileURL, gconf2.Endpoint, "", -1)
	task = &tes.Task{
		Name: "s3 e2e",
		Inputs: []*tes.Input{
			{
				Url:  g1FileURL,
				Path: "/opt/inputs/test-file1.txt",
				Type: tes.FileType_FILE,
			},
			{
				Url:  g2FileURL,
				Path: "/opt/inputs/test-file2.txt",
				Type: tes.FileType_FILE,
			},
		},
		Outputs: []*tes.Output{
			{
				Path: "/opt/workdir/test-output-file.txt",
				Url:  outFileURL,
				Type: tes.FileType_FILE,
			},
		},
		Executors: []*tes.Executor{
			{
				Image: "alpine:latest",
				Command: []string{
					"sh",
					"-c",
					"cat $(find /opt/inputs -type f | sort) > test-output-file.txt",
				},
				Workdir: "/opt/workdir",
			},
		},
	}
	resp, err = fun.RPC.CreateTask(context.Background(), task)
	if err != nil {
		t.Fatal(err)
	}

	taskFinal = fun.Wait(resp.Id)

	if taskFinal.State != tes.State_SYSTEM_ERROR {
		t.Fatal("expected task failure")
	}
}
