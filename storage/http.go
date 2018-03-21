package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
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
func (b *HTTPBackend) Get(ctx context.Context, rawurl string, hostPath string, class tes.FileType) (err error) {
	switch class {
	case File:
		err := fsutil.EnsurePath(hostPath)
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
		defer src.Body.Close()

		dest, err := os.Create(hostPath)
		if err != nil {
			return err
		}
		defer func() {
			cerr := dest.Close()
			if cerr != nil {
				err = fmt.Errorf("%v; %v", err, cerr)
			}
		}()

		_, err = io.Copy(dest, fsutil.Reader(ctx, src.Body))
		return err

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
		return &ErrUnsupportedProtocol{"httpStorage"}
	}

	if class == Directory {
		return fmt.Errorf("httpStorage: directory file type is not supported")
	}

	resp, err := b.client.Head(rawurl)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("httpStorage: HEAD request to %s returned status: %s", rawurl, resp.Status)
	}
	return nil
}

// SupportsPut indicates whether this backend supports PUT storage requests.
func (b *HTTPBackend) SupportsPut(rawurl string, class tes.FileType) error {
	if !strings.HasPrefix(rawurl, "http://") && !strings.HasPrefix(rawurl, "https://") {
		return &ErrUnsupportedProtocol{"httpStorage"}
	}

	return fmt.Errorf("httpStorage: Put is not supported")
}
