package storage

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"testing"
)

func TestStorageWithConfig(t *testing.T) {
	// Single valid config
	c := config.StorageConfig{
		Local: config.LocalStorage{
			AllowedDirs: []string{"/tmp"},
		},
		GS:        config.GSStorage{Disabled: true},
		AmazonS3:  config.AmazonS3Storage{Disabled: true},
		GenericS3: []config.GenericS3Storage{},
		Swift:     config.SwiftStorage{Disabled: true},
	}
	s := Storage{}
	sc, err := s.WithConfig(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.backends) != 1 {
		t.Fatal("unexpected number of Storage backends")
	}

	// multiple valid configs
	c = config.StorageConfig{
		Local: config.LocalStorage{
			AllowedDirs: []string{"/tmp"},
		},
		GS: config.GSStorage{
			Disabled:    false,
			AccountFile: "",
		},
		AmazonS3: config.AmazonS3Storage{
			Disabled: false,
			AWSConfig: config.AWSConfig{
				Key:    "testkey",
				Secret: "testsecret",
			},
		},
		GenericS3: []config.GenericS3Storage{
			{
				Disabled: false,
				Endpoint: "testendpoint:8080",
				Key:      "testkey",
				Secret:   "testsecret",
			},
		},
		Swift: config.SwiftStorage{Disabled: true},
	}
	sc, err = s.WithConfig(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.backends) != 4 {
		t.Fatal("unexpected number of Storage backends")
	}
}

func TestS3UrlProcessing(t *testing.T) {
	b, err := NewGenericS3Backend(config.GenericS3Storage{
		Endpoint: "s3.amazonaws.com",
		Key:      "",
		Secret:   "",
	})
	if err != nil {
		t.Fatal("Error creating generic S3 backend:", err)
	}

	expectedBucket := "1000genomes"
	expectedKey := "README.analysis_history"

	url := b.parse("s3://s3.amazonaws.com/1000genomes/README.analysis_history")
	if url.bucket != expectedBucket {
		t.Log("expected:", expectedBucket)
		t.Log("actual:", url.bucket)
		t.Error("wrong bucket")
	}
	if url.path != expectedKey {
		t.Log("expected:", expectedKey)
		t.Log("actual:", url.path)
		t.Error("wrong key")
	}

	url = b.parse("s3://1000genomes/README.analysis_history")
	if url.bucket != expectedBucket {
		t.Log("expected:", expectedBucket)
		t.Log("actual:", url.bucket)
		t.Error("wrong bucket")
	}
	if url.path != expectedKey {
		t.Log("expected:", expectedKey)
		t.Log("actual:", url.path)
		t.Error("wrong key")
	}

	ab, err := NewAmazonS3Backend(config.AmazonS3Storage{})
	if err != nil {
		t.Fatal("Error creating generic S3 backend:", err)
	}

	url = ab.parse("s3://s3.amazonaws.com/1000genomes/README.analysis_history")
	if url.bucket != expectedBucket {
		t.Log("expected:", expectedBucket)
		t.Log("actual:", url.bucket)
		t.Error("wrong bucket")
	}
	if url.path != expectedKey {
		t.Log("expected:", expectedKey)
		t.Log("actual:", url.path)
		t.Error("wrong key")
	}

	url = ab.parse("s3://s3.us-west-2.amazonaws.com/1000genomes/README.analysis_history")
	if url.bucket != expectedBucket {
		t.Log("expected:", expectedBucket)
		t.Log("actual:", url.bucket)
		t.Error("wrong bucket")
	}
	if url.path != expectedKey {
		t.Log("expected:", expectedKey)
		t.Log("actual:", url.path)
		t.Error("wrong key")
	}

	url = ab.parse("s3://1000genomes/README.analysis_history")
	if url.bucket != expectedBucket {
		t.Log("expected:", expectedBucket)
		t.Log("actual:", url.bucket)
		t.Error("wrong bucket")
	}
	if url.path != expectedKey {
		t.Log("expected:", expectedKey)
		t.Log("actual:", url.path)
		t.Error("wrong key")
	}
}
