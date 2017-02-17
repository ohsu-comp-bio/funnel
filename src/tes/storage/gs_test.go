package storage

import (
	"os"
	"tes"
	"testing"
)

func authed(t *testing.T) *GSBackend {
	accountFile := os.Getenv("TES_TEST_GS_ACCOUNT_FILE")

	if accountFile == "" {
		t.Skip("No Google Cloud account file. Set TES_TEST_GS_ACCOUNT_FILE")
	}
	conf := tes.GoogleStorageConfig{
		AccountFile: accountFile,
	}

	var err error
	var gs *GSBackend
	gs, err = NewGSBackend(conf)
	if err != nil {
		t.Errorf("Can't get authed backend: %s", err)
	}
	return gs
}

func TestAnonymousGet(t *testing.T) {
	conf := tes.GoogleStorageConfig{}
	gs, err := NewGSBackend(conf)
	if err != nil {
		t.Error(err)
	}

	// TODO this doesn't create the output path yet
	gerr := gs.Get("gs://uspto-pair/applications/05900016.zip", "_test_download/05900016.zip", "File")
	if gerr != nil {
		t.Error(gerr)
	}
}

func TestGet(t *testing.T) {
	gs := authed(t)

	gerr := gs.Get("gs://uspto-pair/applications/05900016.zip", "_test_download/downloaded", "File")
	if gerr != nil {
		t.Error(gerr)
	}
}

func TestPut(t *testing.T) {
	gs := authed(t)

	gerr := gs.Put("gs://ohsu-cromwell-testing.appspot.com/go_test_put", "_test_files/for_put", "File")
	if gerr != nil {
		t.Error(gerr)
	}
}

func TestTrimSlashes(t *testing.T) {
	gs := authed(t)

	gerr := gs.Put("gs://ohsu-cromwell-testing.appspot.com///go_test_put", "_test_files/for_put", "File")
	if gerr != nil {
		t.Error(gerr)
	}
}
