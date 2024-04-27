package storage

import (
	"context"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ohsu-comp-bio/funnel/config"
	"google.golang.org/api/option"
	"google.golang.org/api/storage/v1"
)

func TestGenericS3AnonymousGet(t *testing.T) {
	store, err := NewGenericS3(config.GenericS3Storage{
		Endpoint: "https://s3.amazonaws.com/",
		Key:      "",
		Secret:   "",
	})
	if err != nil {
		t.Fatal("Error creating generic S3 backend:", err)
	}

	_, err = store.Get(context.Background(), "s3://1000genomes/README.analysis_history", "_test_download/README.analysis_history")
	if err != nil {
		t.Error("Error downloading file:", err)
	}
}

func TestAmazonS3AnonymousGet(t *testing.T) {
	c := aws.NewConfig().WithCredentials(credentials.AnonymousCredentials).WithMaxRetries(10)
	s, err := session.NewSession(c)
	if err != nil {
		t.Fatal("Error creating amazon S3 backend:", err)
	}

	store := &AmazonS3{
		sess:     s,
		endpoint: "",
	}

	_, err = store.Get(context.Background(), "s3://1000genomes/README.analysis_history", "_test_download/README.analysis_history")
	if err != nil {
		t.Error("Error downloading file:", err)
	}
}

func TestGoogleStorageAnonymousGet(t *testing.T) {
	svc, err := storage.NewService(context.TODO(), option.WithHTTPClient(&http.Client{}))
	if err != nil {
		t.Fatal("Error creating GS backend:", err)
	}

	store := &GoogleCloud{svc}

	_, err = store.Get(context.Background(), "gs://uspto-pair/applications/07820856.zip", "_test_download/07820856.zip")
	if err != nil {
		t.Error("Error downloading file:", err)
	}
}
