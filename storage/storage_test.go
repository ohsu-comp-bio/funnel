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
		GS:       config.GSStorage{Disabled: true},
		AmazonS3: config.AmazonS3Storage{Disabled: true},
		S3:       []config.S3Storage{},
		Swift:    config.SwiftStorage{Disabled: true},
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
			AWS: config.AWSConfig{
				Key:    "testkey",
				Secret: "testsecret",
			},
		},
		S3: []config.S3Storage{
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
	b, err := NewGenericS3Backend(config.S3Storage{
		Endpoint: "s3.amazonaws.com",
		Key:      "",
		Secret:   "",
	})
	if err != nil {
		t.Fatal("Error creating generic S3 backend:", err)
	}

	expectedBucket := "1000genomes"
	expectedKey := "README.analysis_history"

	bucket, key := b.processUrl("s3://s3.amazonaws.com/1000genomes/README.analysis_history")
	if bucket != expectedBucket {
		t.Log("expected:", expectedBucket)
		t.Log("actual:", bucket)
		t.Error("wrong bucket")
	}
	if key != expectedKey {
		t.Log("expected:", expectedKey)
		t.Log("actual:", key)
		t.Error("wrong key")
	}

	bucket, key = b.processUrl("s3://1000genomes/README.analysis_history")
	if bucket != expectedBucket {
		t.Log("expected:", expectedBucket)
		t.Log("actual:", bucket)
		t.Error("wrong bucket")
	}
	if key != expectedKey {
		t.Log("expected:", expectedKey)
		t.Log("actual:", key)
		t.Error("wrong key")
	}

	ab, err := NewAmazonS3Backend(config.AmazonS3Storage{})
	if err != nil {
		t.Fatal("Error creating generic S3 backend:", err)
	}

	bucket, key = ab.processUrl("s3://s3.amazonaws.com/1000genomes/README.analysis_history")
	if bucket != expectedBucket {
		t.Log("expected:", expectedBucket)
		t.Log("actual:", bucket)
		t.Error("wrong bucket")
	}
	if key != expectedKey {
		t.Log("expected:", expectedKey)
		t.Log("actual:", key)
		t.Error("wrong key")
	}

	bucket, key = ab.processUrl("s3://s3.us-west-2.amazonaws.com/1000genomes/README.analysis_history")
	if bucket != expectedBucket {
		t.Log("expected:", expectedBucket)
		t.Log("actual:", bucket)
		t.Error("wrong bucket")
	}
	if key != expectedKey {
		t.Log("expected:", expectedKey)
		t.Log("actual:", key)
		t.Error("wrong key")
	}

	bucket, key = ab.processUrl("s3://1000genomes/README.analysis_history")
	if bucket != expectedBucket {
		t.Log("expected:", expectedBucket)
		t.Log("actual:", bucket)
		t.Error("wrong bucket")
	}
	if key != expectedKey {
		t.Log("expected:", expectedKey)
		t.Log("actual:", key)
		t.Error("wrong key")
	}
}
