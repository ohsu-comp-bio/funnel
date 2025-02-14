package storage

import (
	"context"
	"fmt"
	urllib "net/url"
	"os"
	"strings"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/storage/htsget"
)

// HTSGET provides read-access to public URLs.
// It is a client implementation based on the specification
// http://samtools.github.io/hts-specs/htsget.html
//
// HTSGET URLs need to provided in Funnel tasks as
// `htsget://{reads|variants}/resource/id`
//
// Optionally, a Bearer token or 'username:password' can be specified at the
// end of the URL right after the hash-sign to forward credentials.
// 1. `htsget://{reads|variants}/resource/id#basic-user:pass`
// 2. `htsget://{reads|variants}/resource/id#bearer-token`
//
// If credentials are omitted and the request for creating the task contains a
// Bearer token, it will be automatically appended to the URL by Funnel.
type HTSGET struct {
	conf config.HTSGETStorage
}

// NewHTSGET creates a new HTSGET instance based on the provided configuration.
func NewHTSGET(conf config.HTSGETStorage) (*HTSGET, error) {
	return &HTSGET{conf: conf}, nil
}

// Join a directory URL with a subpath. Not supported with HTSGET.
func (b *HTSGET) Join(url, path string) (string, error) {
	return "", fmt.Errorf("htsgetStorage: Join operation is not supported")
}

// Stat returns information about the object at the given storage URL. Not supported with HTSGET.
func (b *HTSGET) Stat(ctx context.Context, url string) (*Object, error) {
	return nil, fmt.Errorf("htsgetStorage: Stat operation is not supported")
}

// List a directory. Calling List on a File is an error. Not supported with HTSGET.
func (b *HTSGET) List(ctx context.Context, url string) ([]*Object, error) {
	return nil, fmt.Errorf("htsgetStorage: List operation is not supported")
}

// Not supported with HTSGET.
func (b *HTSGET) Put(ctx context.Context, url, path string) (*Object, error) {
	return nil, fmt.Errorf("htsgetStorage: Put operation is not supported")
}

// Get operation copies a file from a given URL to the host path.
//
// If configuration specifies sending a public key, the received content will
// be also decrypted locally before writing to the file.
func (b *HTSGET) Get(ctx context.Context, url, path string) (*Object, error) {
	httpsUrl, cleanHtsgetUrl, token, err := b.resolveUrl(url)
	if err != nil {
		return nil, err
	}

	client := htsget.NewClient(httpsUrl, token, time.Duration(b.conf.Timeout))
	err = client.DownloadTo(path)
	if err != nil {
		return nil, err
	}

	// Check that the destination file exists:
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	return &Object{
		URL:          cleanHtsgetUrl,
		Name:         path,
		Size:         info.Size(),
		LastModified: info.ModTime(),
	}, nil
}

// UnsupportedOperations describes which operations (Get, Put, etc) are not
// supported for the given URL.
func (b *HTSGET) UnsupportedOperations(url string) UnsupportedOperations {
	if err := b.supportsPrefix(url); err != nil {
		return AllUnsupported(err)
	}

	ops := UnsupportedOperations{
		List: fmt.Errorf("htsgetStorage: List operation is not supported"),
		Put:  fmt.Errorf("htsgetStorage: Put operation is not supported"),
		Join: fmt.Errorf("htsgetStorage: Join operation is not supported"),
		Stat: fmt.Errorf("htsgetStorage: Stat operation is not supported"),
	}
	return ops
}

func (b *HTSGET) supportsPrefix(url string) error {
	if !strings.HasPrefix(url, "htsget://") {
		return &ErrUnsupportedProtocol{"htsgetStorage"}
	} else if !strings.HasPrefix(url, "htsget://variants/") &&
		!strings.HasPrefix(url, "htsget://reads/") {
		return &ErrInvalidURL{"htsgetStorage"}
	}
	return nil
}

func (b *HTSGET) resolveUrl(htsgetUrl string) (httpUrl string, cleanHtsgetUrl string, token string, err error) {
	// Extract optional explicit authentication credentials
	// (token or user:pass) after the last '#':
	if pos := strings.LastIndex(htsgetUrl, "#"); pos > 0 {
		token = htsgetUrl[pos+1:]
		htsgetUrl = htsgetUrl[:pos]
	}

	// The URL is based on the configured ServiceURL (with path appended):
	actual, err := urllib.Parse(b.conf.ServiceURL)
	if err != nil {
		return
	}

	// Apply the (optional) user:pass to the constructed URL:
	if user, pass, found := strings.Cut(token, ":"); found {
		actual.User = urllib.UserPassword(user, pass)
		token = ""
	}

	// Extract possible query paramaters – they would not go to path:
	if prefix, query, found := strings.Cut(htsgetUrl, "?"); found {
		actual.RawQuery = query
		htsgetUrl = prefix
	}

	// Append the provided path:
	path := htsgetUrl[len("htsget:/"):] // Note: path will begin with "/".
	actual.Path = strings.TrimSuffix(actual.Path, "/") + path

	httpUrl = actual.String()
	cleanHtsgetUrl = htsgetUrl
	return httpUrl, cleanHtsgetUrl, token, nil
}
