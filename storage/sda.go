package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	urllib "net/url"
	"os"
	"strings"
	"time"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/storage/crypt4gh"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

// SDA (sensitive data archive) provides read-access to public URLs.
// SDA URLs need to provided in Funnel tasks as
// `sda://dataset-id/path/to/dataset/file.c4gh`
// The Bearer token is implicitly taken from a task request and used when
// requesting a file from SDA.
type SDA struct {
	conf   config.SDAStorage
	client *http.Client
	log    *logger.Logger
}

// NewSDA creates a new SDA-client instance based on the provided configuration.
func NewSDA(conf config.SDAStorage) (*SDA, error) {
	client := &http.Client{
		Timeout: time.Duration(conf.Timeout),
	}
	log := logger.NewLogger("htsget", logger.DefaultConfig())
	return &SDA{conf, client, log}, nil
}

// UnsupportedOperations describes which operations (Get, Put, etc) are not
// supported for the given URL.
func (b *SDA) UnsupportedOperations(url string) UnsupportedOperations {
	if !strings.HasPrefix(url, "sda://") {
		return AllUnsupported(&ErrUnsupportedProtocol{"sdaStorage"})
	}
	return UnsupportedOperations{
		Join: fmt.Errorf("sdaStorage: Join operation is not supported"),
		List: fmt.Errorf("sdaStorage: List operation is not supported"),
		Stat: fmt.Errorf("sdaStorage: Stat operation is not supported"),
		Put:  fmt.Errorf("sdaStorage: Put operation is not supported"),
	}
}

// Join a directory URL with a subpath. Not supported with SDA.
func (b *SDA) Join(url, path string) (string, error) {
	return "", fmt.Errorf("sdaStorage: Join operation is not supported")
}

// List a directory. Calling List on a File is an error. Not supported with SDA.
func (b *SDA) List(ctx context.Context, url string) ([]*Object, error) {
	return nil, fmt.Errorf("sdaStorage: List operation is not supported")
}

// Not supported with SDA.
func (b *SDA) Put(ctx context.Context, url, path string) (*Object, error) {
	return nil, fmt.Errorf("sdaStorage: Put operation is not supported")
}

// Stat returns information about the object in SDA.
func (s *SDA) Stat(ctx context.Context, url string) (*Object, error) {
	resp, err := s.doRequest(ctx, "HEAD", url, nil)
	if err != nil {
		return nil, err
	}
	return toObject(resp, url), nil
}

// Get operation copies a file from a given URL to the host path.
//
// If configuration specifies sending a public key, the received content will
// be also decrypted locally before writing to the file.
func (s *SDA) Get(ctx context.Context, url, path string) (*Object, error) {
	keys, err := crypt4gh.ResolveKeyPair()
	if err != nil {
		return nil, err
	}

	resp, err := s.doRequest(ctx, "GET", url, keys)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err = downloadToFile(ctx, resp, path, keys); err != nil {
		return nil, err
	}

	return toObject(resp, path), nil
}

func (s *SDA) resolveUrl(sdaUrl string) (httpUrl string, token string, err error) {
	// Extract the Bearer token after the last '#':
	if pos := strings.LastIndex(sdaUrl, "#"); pos > 0 {
		token = sdaUrl[pos+1:]
		sdaUrl = sdaUrl[:pos]
	}

	actual, err := urllib.Parse(s.conf.ServiceURL)
	if err != nil {
		return
	}

	// Extract possible query paramaters – they would not go to path:
	if prefix, query, found := strings.Cut(sdaUrl, "?"); found {
		actual.RawQuery = query
		sdaUrl = prefix
	}

	// Append the provided path (e.g. 'variants/file-path'):
	path := sdaUrl[len("sda://"):]
	if strings.HasSuffix(path, ".c4gh") {
		path = "/s3-encrypted/" + path
	} else {
		path = "/s3/" + path
	}
	actual.Path = strings.TrimSuffix(actual.Path, "/") + path

	httpUrl = actual.String()
	return httpUrl, token, nil
}

func (s *SDA) doRequest(
	ctx context.Context,
	method string,
	sdaUrl string,
	keys *crypt4gh.KeyPair,
) (resp *http.Response, err error) {

	url, token, err := s.resolveUrl(sdaUrl)
	if err != nil {
		return
	}

	s.log.Info("Requesting file: " + url)

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		err = fmt.Errorf("sdaStorage: creating %s request: %s", method, err)
		return
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	if strings.Contains(req.URL.Path, "/s3-encrypted/") && keys != nil {
		s.log.Info("Including the Client-Public-Key header with the Crypt4GH public key.")
		req.Header.Set("Client-Public-Key", keys.EncodePublicKeyBase64())
	}

	resp, err = s.client.Do(req)
	if err != nil {
		err = fmt.Errorf("sdaStorage: executing %s request: %s", method, err)
	} else if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Request.Body)
		resp.Body.Close()
		err = fmt.Errorf("sdaStorage: %s request returned status code %d: %s",
			method, resp.StatusCode, string(body))
	}

	return
}

func downloadToFile(
	ctx context.Context,
	resp *http.Response,
	path string,
	keys *crypt4gh.KeyPair,
) error {
	dest, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("sdaStorage: creating local file: %s", err)
	}
	defer dest.Close()

	var stream io.Reader = resp.Body
	if resp.Request.Header.Get("Client-Public-Key") != "" {
		stream, err = keys.Decrypt(stream)
		if err != nil {
			return fmt.Errorf(
				"sdaStorage: decrypting received payload data: %s", err)
		}
	}

	if _, err = io.Copy(dest, fsutil.Reader(ctx, stream)); err != nil {
		return fmt.Errorf("sdaStorage: copying file: %s", err)
	}

	return nil
}

func toObject(resp *http.Response, path string) *Object {
	modtime, _ := http.ParseTime(resp.Header.Get("Last-Modified"))
	etag := resp.Header.Get("ETag")

	return &Object{
		URL:          resp.Request.URL.String(),
		Name:         path,
		Size:         resp.ContentLength,
		LastModified: modtime,
		ETag:         etag,
	}
}
