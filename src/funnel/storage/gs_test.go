package storage

import (
	"context"
	"funnel/config"
	"funnel/proto/tes"
	"os"
	"testing"
)

func authed(t *testing.T) *GSBackend {
	accountFile := os.Getenv("TES_TEST_GS_ACCOUNT_FILE")

	if accountFile == "" {
		t.Skip("No Google Cloud account file. Set TES_TEST_GS_ACCOUNT_FILE")
	}
	conf := config.GSStorage{
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
	ctx := context.Background()
	conf := config.GSStorage{}
	gs, err := NewGSBackend(conf)
	if err != nil {
		t.Fatal(err)
	}

	gerr := gs.Get(ctx, "gs://uspto-pair/applications/05900016.zip", "_test_download/05900016.zip", tes.FileType_FILE, false)
	if gerr != nil {
		t.Error(gerr)
	}
}

func TestGet(t *testing.T) {
	ctx := context.Background()
	gs := authed(t)

	gerr := gs.Get(ctx, "gs://uspto-pair/applications/05900016.zip", "_test_download/downloaded", tes.FileType_FILE, false)
	if gerr != nil {
		t.Error(gerr)
	}
}

func TestPut(t *testing.T) {
	ctx := context.Background()
	gs := authed(t)

	gerr := gs.Put(ctx, "gs://ohsu-cromwell-testing.appspot.com/go_test_put", "_test_files/for_put", tes.FileType_FILE)
	if gerr != nil {
		t.Error(gerr)
	}
}

func TestTrimSlashes(t *testing.T) {
	ctx := context.Background()
	gs := authed(t)

	gerr := gs.Put(ctx, "gs://ohsu-cromwell-testing.appspot.com///go_test_put", "_test_files/for_put", tes.FileType_FILE)
	if gerr != nil {
		t.Error(gerr)
	}
}
