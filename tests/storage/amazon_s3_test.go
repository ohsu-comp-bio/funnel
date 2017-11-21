package storage

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/storage"
	"github.com/ohsu-comp-bio/funnel/tests"
	util "github.com/ohsu-comp-bio/funnel/util/aws"
	"io/ioutil"
	"testing"
)

func TestAmazonS3Storage(t *testing.T) {
	tests.SetLogOutput(log, t)

	if !conf.Worker.Storage.AmazonS3.Valid() {
		t.Skipf("Skipping amazon s3 e2e tests...")
	}

	testBucket := "funnel-e2e-tests-" + tests.RandomString(6)

	sess, err := util.NewAWSSession(conf.Worker.Storage.AmazonS3.AWS)
	if err != nil {
		t.Fatal("error creating AWS session:", err)
	}
	client := s3.New(sess)

	_, err = client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(testBucket),
	})
	if err != nil {
		t.Fatal("error creating test s3 bucket:", err)
	}
	defer func() {
		amazonEmptyBucket(client, testBucket)
		client.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(testBucket)})
	}()

	protocol := "s3://"

	store, err := storage.Storage{}.WithConfig(conf.Worker.Storage)
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
}

func amazonEmptyBucket(client *s3.S3, bucket string) error {
	params := &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	}
	for {
		//Requesting for batch of objects from s3 bucket
		objects, err := client.ListObjects(params)
		if err != nil {
			return err
		}
		//Checks if the bucket is already empty
		if len((*objects).Contents) == 0 {
			log.Info("Bucket is already empty")
			return nil
		}

		//creating an array of pointers of ObjectIdentifier
		objectsToDelete := make([]*s3.ObjectIdentifier, 0, 1000)
		for _, object := range (*objects).Contents {
			obj := s3.ObjectIdentifier{
				Key: object.Key,
			}
			objectsToDelete = append(objectsToDelete, &obj)
		}
		//Creating JSON payload for bulk delete
		deleteArray := s3.Delete{Objects: objectsToDelete}
		deleteParams := &s3.DeleteObjectsInput{
			Bucket: aws.String(bucket),
			Delete: &deleteArray,
		}
		//Running the Bulk delete job (limit 1000)
		_, err = client.DeleteObjects(deleteParams)
		if err != nil {
			return err
		}
		if *(*objects).IsTruncated { //if there are more objects in the bucket, IsTruncated = true
			params.Marker = (*deleteParams).Delete.Objects[len((*deleteParams).Delete.Objects)-1].Key
		} else { //if all objects in the bucket have been cleaned up.
			break
		}
	}
	return nil
}
