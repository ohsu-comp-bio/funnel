package storage

import (
	"context"
	"testing"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
)

func TestHTTP(t *testing.T) {
	store, err := NewHTTP(config.HTTPStorage{
		Timeout: config.Duration(2 * time.Second),
	})
	if err != nil {
		t.Fatal("Error creating HTTP backend:", err)
	}

	// test client timeout
	err = store.UnsupportedOperations("https://fakeurl.com").Get
	if err == nil {
		t.Error("Expected timeout error")
	}

	// test Get is supported
	err = store.UnsupportedOperations("https://google.com").Get
	if err != nil {
		t.Error("Unexpected error for unsupported.Get call:", err)
	}

	// test Get on file that requires auth
	err = store.UnsupportedOperations("https://s3.amazonaws.com/private").Get
	if err == nil {
		t.Error("Expected error for unsupported.Get call")
	}

	// test Get succeeds
	_, err = store.Get(context.Background(), "https://google.com", "_test_download/test_https_download.html")
	if err != nil {
		t.Error("Unexpected error downloading file:", err)
	}

	// test List not supported
	_, err = store.List(context.Background(), "https://google.com")
	if err == nil {
		t.Error("Expected error for List irectory")
	}

	// test Put not supported
	_, err = store.Put(context.Background(), "https://fakeurl.com", "_test_download/test_https_download.html")
	if err == nil {
		t.Error("Expected error for Put call")
	}
}
