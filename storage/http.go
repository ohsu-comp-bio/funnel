package storage

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
	"io"
	"net/http"
	"os"
	"strings"
)

// HTTPBackend provides read access to public URLs.
type HTTPBackend struct {
	client *http.Client
}

// NewHTTPBackend creates a new HTTPBackend instance.
func NewHTTPBackend(conf config.HTTPStorage) (*HTTPBackend, error) {
	client := &http.Client{
		Timeout: conf.Timeout,
	}
	return &HTTPBackend{client}, nil
}

// Get copies a file from a given URL to the host path.
func (b *HTTPBackend) Get(ctx context.Context, rawurl string, hostPath string, class tes.FileType) error {

	switch class {
	case File:
		fsutil.EnsurePath(hostPath)
		dest, err := os.Create(hostPath)
		if err != nil {
			return err
		}

		req, err := http.NewRequest("GET", rawurl, nil)
		if err != nil {
			return err
		}
		req.WithContext(ctx)

		src, err := b.client.Do(req)
		if err != nil {
			return err
		}

		_, err = io.Copy(dest, src.Body)
		if err != nil {
			return err
		}

		err = src.Body.Close()
		if err != nil {
			return err
		}

		return dest.Close()

	case Directory:
		return fmt.Errorf("Unsupported file class: %s", class)

	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}
}

// PutFile is not supported for the HTTPBackend
func (b *HTTPBackend) PutFile(ctx context.Context, rawurl string, hostPath string) error {
	return fmt.Errorf("PutFile - Not Supported")
}

// SupportsGet indicates whether this backend supports GET storage requests.
func (b *HTTPBackend) SupportsGet(rawurl string, class tes.FileType) error {
	if !strings.HasPrefix(rawurl, "http://") && !strings.HasPrefix(rawurl, "https://") {
		return fmt.Errorf("http(s): unsupported protocol; expected http:// or https://")
	}
	if class == Directory {
		return fmt.Errorf("http(s): directory file type is not supported")
	}
	return nil
}

// SupportsPut indicates whether this backend supports PUT storage requests.
func (b *HTTPBackend) SupportsPut(rawurl string, class tes.FileType) error {
	return fmt.Errorf("http(s): Put is not supported")
}
