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
			Disabled: false,
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
