package storage

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
)

func TestStorageWithConfig(t *testing.T) {
	// Single valid config
	c := config.Config{
		LocalStorage: config.LocalStorage{
			AllowedDirs: []string{"/tmp"},
		},
		GoogleStorage: config.GSStorage{Disabled: true},
		AmazonS3:      config.AmazonS3Storage{Disabled: true},
		GenericS3:     []config.GenericS3Storage{},
		Swift:         config.SwiftStorage{Disabled: true},
		HTTPStorage:   config.HTTPStorage{Disabled: true},
	}

	sc, err := NewStorage(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.backends) != 1 {
		t.Fatal("unexpected number of Storage backends")
	}

	// multiple valid configs
	c = config.Config{
		LocalStorage: config.LocalStorage{
			AllowedDirs: []string{"/tmp"},
		},
		GoogleStorage: config.GSStorage{
			Disabled:        false,
			CredentialsFile: "",
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
				Endpoint: "http://testendpoint:8080",
				Key:      "testkey",
				Secret:   "testsecret",
			},
		},
		Swift: config.SwiftStorage{
			Disabled:   false,
			UserName:   "fakeuser",
			Password:   "fakepassword",
			AuthURL:    "http://testendpoint:5000/v2.0",
			TenantName: "faketenantname",
			TenantID:   "faketenantid",
			RegionName: "fakeregion",
		},
		HTTPStorage: config.HTTPStorage{Disabled: false},
	}
	sc, err = NewStorage(c)
	if err != nil {
		t.Fatal(err)
	}
	if len(sc.backends) != 6 {
		t.Fatal("unexpected number of Storage backends")
	}
}

func TestUrlParsing(t *testing.T) {
	expectedBucket := "1000genomes"
	expectedKey := "README.analysis_history"

	// Generic S3
	b, err := NewGenericS3Backend(config.GenericS3Storage{
		Endpoint: "s3.amazonaws.com",
	})
	if err != nil {
		t.Error("Error creating generic S3 backend:", err)
	}

	url, err := b.parse("s3://s3.amazonaws.com/1000genomes/README.analysis_history")
	if err != nil {
		t.Error("unexpected error", err)
	}
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

	url, err = b.parse("s3://1000genomes/README.analysis_history")
	if err != nil {
		t.Error("unexpected error", err)
	}
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

	url, err = b.parse("gs://1000genomes/README.analysis_history")
	if _, ok := err.(*ErrUnsupportedProtocol); !ok {
		t.Error("expected ErrUnsupportedProtocol")
	}

	url, err = b.parse("s3://")
	if _, ok := err.(*ErrInvalidURL); !ok {
		t.Error("expected ErrInvalidURL")
	}

	// Amazon S3
	ab, err := NewAmazonS3Backend(config.AmazonS3Storage{})
	if err != nil {
		t.Error("Error creating amazon S3 backend:", err)
	}

	url, err = ab.parse("s3://s3.amazonaws.com/1000genomes/README.analysis_history")
	if err != nil {
		t.Error("unexpected error", err)
	}
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

	url, err = ab.parse("s3://s3.us-west-2.amazonaws.com/1000genomes/README.analysis_history")
	if err != nil {
		t.Error("unexpected error", err)
	}
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

	url, err = ab.parse("s3://1000genomes/README.analysis_history")
	if err != nil {
		t.Error("unexpected error", err)
	}
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

	url, err = ab.parse("gs://1000genomes/README.analysis_history")
	if _, ok := err.(*ErrUnsupportedProtocol); !ok {
		t.Error("expected ErrUnsupportedProtocol")
	}

	url, err = ab.parse("s3://")
	if _, ok := err.(*ErrInvalidURL); !ok {
		t.Error("expected ErrInvalidURL")
	}

	// Google Storage
	gb, err := NewGSBackend(config.GSStorage{})
	if err != nil {
		t.Error("Error creating google storage backend:", err)
	}

	url, err = gb.parse("gs://1000genomes/README.analysis_history")
	if err != nil {
		t.Error("unexpected error", err)
	}
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

	url, err = gb.parse("s3://1000genomes/README.analysis_history")
	if _, ok := err.(*ErrUnsupportedProtocol); !ok {
		t.Error("expected ErrUnsupportedProtocol")
	}

	url, err = gb.parse("gs://")
	if _, ok := err.(*ErrInvalidURL); !ok {
		t.Error("expected ErrInvalidURL")
	}

	// Swift
	sb := &SwiftBackend{}

	url, err = sb.parse("swift://1000genomes/README.analysis_history")
	if err != nil {
		t.Error("unexpected error", err)
	}
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

	url, err = sb.parse("s3://1000genomes/README.analysis_history")
	if _, ok := err.(*ErrUnsupportedProtocol); !ok {
		t.Error("expected ErrUnsupportedProtocol")
	}

	url, err = sb.parse("swift://")
	if _, ok := err.(*ErrInvalidURL); !ok {
		t.Error("expected ErrInvalidURL")
	}
}

func TestWalkFiles(t *testing.T) {
	tmp, err := ioutil.TempDir("", "funnel-test-local-storage")
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(path.Join(tmp, "test_file"))
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	files, err := walkFiles(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 {
		t.Fatal("unexpected number of files returned by walkFiles")
	}

	nonexistent := path.Join(tmp, "this/path/doesnt/exist")
	_, err = walkFiles(nonexistent)
	if err == nil {
		t.Fatal("expected error")
	}
}
