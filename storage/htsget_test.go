package storage

import (
	"context"
	"testing"

	"github.com/ohsu-comp-bio/funnel/config"
)

func TestHTSGET(t *testing.T) {
	invalidUrl := "https://example.org"
	validUrls := []string{"htsget://reads/file", "htsget://variants/file"}

	store, err := NewHTSGET(config.HTSGETStorage{})
	if err != nil {
		t.Fatal("Unexpected error while creating an HTSGET backend:", err)
	}

	// Wrong protocol results in unsupported operation for each action
	ops := store.UnsupportedOperations(invalidUrl)
	if ops.Stat == nil || ops.Get == nil || ops.Join == nil || ops.List == nil || ops.Put == nil {
		t.Error("Not all operations were denied when an HTTPS URL was specified to HTSGET")
	}

	// Correct protocol results in some unsupported operations (except GET)
	for _, validUrl := range validUrls {
		ops = store.UnsupportedOperations(validUrl)

		if ops.Stat == nil || ops.Join == nil || ops.List == nil || ops.Put == nil {
			t.Error("Some non-supported operations were permitted for an HTSGET URL")
		} else if ops.Get != nil {
			t.Error("GET operation was not permitted for an HTSGET URL", err)
		}

		// Verifying unsupported operations

		if _, err = store.Stat(context.Background(), validUrl); err == nil {
			t.Error("Stat call should have failed for HTSGET")
		}

		if _, err = store.Join(validUrl, ""); err == nil {
			t.Error("Join call should have failed for HTSGET")
		}

		if _, err = store.List(context.Background(), validUrl); err == nil {
			t.Error("List call should have failed for HTSGET")
		}

		if _, err = store.Put(context.Background(), validUrl, "path"); err == nil {
			t.Error("Put call should have failed for HTSGET")
		}
	}
}
