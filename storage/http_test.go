package storage

import (
	"context"
	"github.com/ohsu-comp-bio/funnel/config"
	"testing"
	"time"
)

func TestHTTPBackend(t *testing.T) {
	b, err := NewHTTPBackend(config.HTTPStorage{
		Timeout: 2 * time.Second,
	})
	if err != nil {
		t.Fatal("Error creating HTTP backend:", err)
	}
	store := Storage{}.WithBackend(b)

	err = store.Get(context.Background(), "https://google.com", "_test_download/test_https_download.html", File)
	if err != nil {
		t.Error("Error downloading file:", err)
	}

	err = store.Get(context.Background(), "https://google.com", "_test_download/test_https_download.html", Directory)
	if err == nil {
		t.Error("Expected error")
	}

	_, err = store.Put(context.Background(), "https://fakeurl.com", "_test_download/test_https_download.html", File)
	if err == nil {
		t.Error("Expected error")
	}
}
