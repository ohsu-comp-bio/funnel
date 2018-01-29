package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

func TestRetrier(t *testing.T) {
	f := &fakeBackend{}
	r := retrier{backend: f, maxTries: 100}
	bg := context.Background()

	// Test that get is retried
	// Fail twice, then succeed.
	f.onGet = func() error {
		if f.getCalls < 3 {
			return fmt.Errorf("should retry")
		}
		return nil
	}
	r.Get(bg, "url", "path", tes.FileType_FILE)
	if f.getCalls != 3 {
		t.Errorf("expected Get to be called 3 times, got %d", f.getCalls)
	}

	// Test that a successful call is not retried
	r.SupportsPut("url", tes.FileType_FILE)
	if f.supportsPutCalls != 1 {
		t.Errorf("expected SupportsPut to be called 1 times, got %d", f.supportsPutCalls)
	}

	// Test shouldRetry function
	errSupportsGet := fmt.Errorf("supports get error")
	f.onSupportsGet = func() error {
		return errSupportsGet
	}
	r.shouldRetry = func(err error) bool {
		if err != errSupportsGet {
			t.Error("expected error value to be passed to shouldRetry")
		}
		return false
	}
	serr := r.SupportsGet("url", tes.FileType_FILE)
	if f.supportsGetCalls != 1 {
		t.Error("expected SupportsGet to be called only once")
	}
	if serr != errSupportsGet {
		t.Error("expected SupportsGet error to be passed through")
	}

	// Reset should retry
	r.shouldRetry = nil
}

func TestRetrierMaxTries(t *testing.T) {
	f := &fakeBackend{}
	r := retrier{backend: f, maxTries: 5}
	bg := context.Background()

	f.onGet = func() error {
		return fmt.Errorf("always fail")
	}

	r.Get(bg, "url", "path", tes.FileType_FILE)
	if f.getCalls != 5 {
		t.Errorf("expected Get to be called 5 times, got %d", f.getCalls)
	}
}

type fakeBackend struct {
	getCalls, putCalls, supportsGetCalls, supportsPutCalls int
	onGet, onPut, onSupportsGet, onSupportsPut             func() error
}

func (f *fakeBackend) Get(ctx context.Context, url, path string, class tes.FileType) error {
	f.getCalls++
	if f.onGet != nil {
		return f.onGet()
	}
	return nil
}
func (f *fakeBackend) PutFile(ctx context.Context, url, path string) error {
	f.putCalls++
	if f.onPut != nil {
		return f.onPut()
	}
	return nil
}
func (f *fakeBackend) SupportsGet(url string, class tes.FileType) error {
	f.supportsGetCalls++
	if f.onSupportsGet != nil {
		return f.onSupportsGet()
	}
	return nil
}
func (f *fakeBackend) SupportsPut(url string, class tes.FileType) error {
	f.supportsPutCalls++
	if f.onSupportsPut != nil {
		return f.onSupportsPut()
	}
	return nil
}
