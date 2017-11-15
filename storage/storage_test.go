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
		GS:    []config.GSStorage{},
		S3:    config.S3Storage{},
		Swift: config.SwiftStorage{Disabled: true},
	}
	s := Storage{}
	sc, err := s.WithConfig(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.backends) != 2 {
		t.Fatal("unexpected number of Storage backends")
	}

	// multiple valid configs
	c = config.StorageConfig{
		Local: config.LocalStorage{
			AllowedDirs: []string{"/tmp"},
		},
		GS: []config.GSStorage{
			{
				FromEnv: true,
			},
		},
		S3: config.S3Storage{
			Credentials: config.AWSCredentials{
				Key:    "testkey",
				Secret: "testsecret",
			},
		},
		Swift: config.SwiftStorage{Disabled: true},
	}
	sc, err = s.WithConfig(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.backends) != 3 {
		t.Fatal("unexpected number of Storage backends")
	}
}
