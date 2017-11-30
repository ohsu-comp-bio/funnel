package storage

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/ohsu-comp-bio/funnel/config"
	"google.golang.org/api/storage/v1"
	"net/http"
	"testing"
)

func TestGenericS3AnonymousGet(t *testing.T) {
	b, err := NewGenericS3Backend(config.GenericS3Storage{
		Endpoint: "https://s3.amazonaws.com/",
		Key:      "",
		Secret:   "",
	})
	if err != nil {
		t.Fatal("Error creating generic S3 backend:", err)
	}
	store := Storage{}.WithBackend(b)

	err = store.Get(context.Background(), "s3://1000genomes/README.analysis_history", "_test_download/README.analysis_history", File)
	if err != nil {
		t.Error("Error downloading file:", err)
	}
}

func TestAmazonS3AnonymousGet(t *testing.T) {
	c := aws.NewConfig().WithCredentials(credentials.AnonymousCredentials)
	s, err := session.NewSession(c)
	if err != nil {
		t.Fatal("Error creating amazon S3 backend:", err)
	}

	store := Storage{}.WithBackend(&AmazonS3Backend{
		sess:     s,
		endpoint: "",
	})

	err = store.Get(context.Background(), "s3://1000genomes/README.analysis_history", "_test_download/README.analysis_history", File)
	if err != nil {
		t.Error("Error downloading file:", err)
	}
}

func TestGoogleStorageAnonymousGet(t *testing.T) {
	svc, err := storage.New(&http.Client{})
	if err != nil {
		t.Fatal("Error creating GS backend:", err)
	}

	store := Storage{}.WithBackend(&GSBackend{svc})

	err = store.Get(context.Background(), "gs://uspto-pair/applications/05900016.zip", "_test_download/05900016.zip", File)
	if err != nil {
		t.Error("Error downloading file:", err)
	}
}
