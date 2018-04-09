package storage

import (
	"context"
	"testing"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
)

func TestHTTPBackend(t *testing.T) {
	b, err := NewHTTPBackend(config.HTTPStorage{
		Timeout: config.Duration(2 * time.Second),
	})
	if err != nil {
		t.Fatal("Error creating HTTP backend:", err)
	}
	store := Storage{}.WithBackend(b)

	// test client timeout
	err = store.SupportsGet("https://fakeurl.com", File)
	if err == nil {
		t.Error("Expected timeout error")
	}

	// test SupportsGet
	err = store.SupportsGet("https://google.com", File)
	if err != nil {
		t.Error("Unexpected error for SupportsGet call:", err)
	}

	// test SupportsGet on file that requires auth
	err = store.SupportsGet("https://s3.amazonaws.com/private", File)
	if err == nil {
		t.Error("Expected error for SupportsGet call")
	}

	// test Get succeeds
	err = store.Get(context.Background(), "https://google.com", "_test_download/test_https_download.html", File)
	if err != nil {
		t.Error("Unexpected error downloading file:", err)
	}

	// test Get fails for Directory
	err = store.Get(context.Background(), "https://google.com", "_test_download/test_https_download.html", Directory)
	if err == nil {
		t.Error("Expected error for Get call on Directory")
	}

	// test supportsPut
	_, err = store.Put(context.Background(), "https://fakeurl.com", "_test_download/test_https_download.html", File)
	if err == nil {
		t.Error("Expected error for Put call")
	}
}
