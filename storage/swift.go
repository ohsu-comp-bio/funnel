package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alecthomas/units"
	"github.com/ncw/swift"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

const swiftProtocol = "swift://"

// Swift provides access to an sw object store.
type Swift struct {
	conn      *swift.Connection
	chunkSize int64
}

// NewSwift creates an Swift client instance, give an endpoint URL
// and a set of authentication credentials.
func NewSwift(conf config.SwiftStorage) (*Swift, error) {

	// Create a connection
	conn := &swift.Connection{
		UserName: conf.UserName,
		ApiKey:   conf.Password,
		AuthUrl:  conf.AuthURL,
		Tenant:   conf.TenantName,
		TenantId: conf.TenantID,
		Region:   conf.RegionName,
	}

	// Read environment variables and apply them to the Connection structure.
	// Won't overwrite any parameters which are already set in the Connection struct.
	err := conn.ApplyEnvironment()
	if err != nil {
		return nil, err
	}

	var chunkSize int64
	if conf.ChunkSizeBytes < int64(100*units.MB) {
		chunkSize = int64(500 * units.MB)
	} else if conf.ChunkSizeBytes > int64(5*units.GB) {
		chunkSize = int64(5 * units.GB)
	} else {
		chunkSize = conf.ChunkSizeBytes
	}

	return &Swift{conn, chunkSize}, nil
}

// NewSwiftRetrier returns a Swift storage client that retries operations on error.
func NewSwiftRetrier(conf config.SwiftStorage) (*Retrier, error) {
	b, err := NewSwift(conf)
	if err != nil {
		return nil, err
	}
	return &Retrier{
		Backend: b,
		Retrier: &util.Retrier{
			MaxTries:            conf.MaxRetries,
			InitialInterval:     time.Second * 5,
			MaxInterval:         time.Minute * 5,
			Multiplier:          2.0,
			RandomizationFactor: 0.5,
			MaxElapsedTime:      0,
			ShouldRetry: func(err error) bool {
				// Retry on errors that swift names specifically.
				if err == swift.ObjectCorrupted || err == swift.TimeoutError {
					return true
				}
				// Retry on service unavailable.
				if se, ok := err.(*swift.Error); ok {
					return se.StatusCode == http.StatusServiceUnavailable
				}
				return false
			},
		},
	}, nil
}

// Stat returns metadata about the given url, such as checksum.
func (sw *Swift) Stat(ctx context.Context, url string) (*Object, error) {
	u, err := sw.parse(url)
	if err != nil {
		return nil, err
	}

	info, _, err := sw.conn.Object(u.bucket, u.path)
	if err != nil {
		return nil, fmt.Errorf("getting object info: %s", err)
	}
	return &Object{
		URL:          url,
		Name:         info.Name,
		Size:         info.Bytes,
		LastModified: info.LastModified,
		ETag:         info.Hash,
	}, nil
}

// List lists the objects at the given url.
func (sw *Swift) List(ctx context.Context, url string) ([]*Object, error) {
	u, err := sw.parse(url)
	if err != nil {
		return nil, err
	}

	objs, err := sw.conn.ObjectsAll(u.bucket, &swift.ObjectsOpts{
		Prefix: u.path,
	})
	if err != nil {
		return nil, fmt.Errorf("listing objects by prefix: %s", err)
	}

	var objects []*Object
	for _, obj := range objs {
		objects = append(objects, &Object{
			URL:          swiftProtocol + u.bucket + "/" + obj.Name,
			Name:         obj.Name,
			Size:         obj.Bytes,
			LastModified: obj.LastModified,
			ETag:         obj.Hash,
		})
	}
	return objects, nil
}

// Get copies an object from storage to the host path.
func (sw *Swift) Get(ctx context.Context, url, path string) (obj *Object, err error) {
	u, err := sw.parse(url)
	if err != nil {
		return nil, err
	}

	var checkHash = true
	var headers swift.Headers

	obj, err = sw.Stat(ctx, url)
	if err != nil {
		return
	}

	f, _, err := sw.conn.ObjectOpen(u.bucket, u.path, checkHash, headers)
	if err != nil {
		err = fmt.Errorf("initiating download: %s", err)
		return
	}
	defer func() {
		cerr := f.Close()
		if cerr != nil {
			err = fmt.Errorf("closing file %v; %v", err, cerr)
		}
	}()

	dest, err := os.Create(path)
	if err != nil {
		err = fmt.Errorf("creating file: %s", err)
	}
	defer func() {
		cerr := dest.Close()
		if cerr != nil {
			err = fmt.Errorf("%v; %v", err, cerr)
		}
	}()

	_, err = io.Copy(dest, fsutil.Reader(ctx, f))
	if err != nil {
		err = fmt.Errorf("copying file: %s", err)
		return
	}

	return
}

// Put copies an object (file) from the host path to storage.
func (sw *Swift) Put(ctx context.Context, url, path string) (*Object, error) {

	u, err := sw.parse(url)
	if err != nil {
		return nil, err
	}

	reader, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening host file %q: %s", path, err)
	}
	defer reader.Close()

	var writer io.WriteCloser
	var checkHash = true
	var hash string
	var contentType string
	var headers swift.Headers

	fSize := fsutil.FileSize(path)
	if fSize < int64(5*units.GB) {
		writer, err = sw.conn.ObjectCreate(u.bucket, u.path, checkHash, hash, contentType, headers)
	} else {
		writer, err = sw.conn.StaticLargeObjectCreateFile(&swift.LargeObjectOpts{
			Container:  u.bucket,
			ObjectName: u.path,
			CheckHash:  checkHash,
			Hash:       hash,
			Headers:    headers,
			ChunkSize:  sw.chunkSize,
		})
	}

	if err != nil {
		return nil, fmt.Errorf("creating object: %s", err)
	}

	_, copyErr := io.Copy(writer, fsutil.Reader(ctx, reader))
	// In order to do the Stat call below, the writer needs to be closed
	// so that the object is created.
	closeErr := writer.Close()

	if copyErr != nil {
		return nil, fmt.Errorf("copying file: %s", copyErr)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("closing file: %s", closeErr)
	}

	return sw.Stat(ctx, url)
}

// Join joins the given URL with the given subpath.
func (sw *Swift) Join(url, path string) (string, error) {
	return strings.TrimSuffix(url, "/") + "/" + path, nil
}

// UnsupportedOperations describes which operations (Get, Put, etc) are not
// supported for the given URL.
func (sw *Swift) UnsupportedOperations(url string) UnsupportedOperations {
	u, err := sw.parse(url)
	if err != nil {
		return AllUnsupported(err)
	}
	_, _, err = sw.conn.Container(u.bucket)
	if err != nil {
		return AllUnsupported(fmt.Errorf("swift: failed to find bucket: %s. error: %v", u.bucket, err))
	}
	return AllSupported()
}

func (sw *Swift) parse(rawurl string) (*urlparts, error) {
	ok := strings.HasPrefix(rawurl, swiftProtocol)
	if !ok {
		return nil, &ErrUnsupportedProtocol{"swift"}
	}

	path := strings.TrimPrefix(rawurl, swiftProtocol)
	if path == "" {
		return nil, &ErrInvalidURL{"swift"}
	}

	split := strings.SplitN(path, "/", 2)
	url := &urlparts{}
	if len(split) > 0 {
		url.bucket = split[0]
	}
	if len(split) == 2 {
		url.path = split[1]
	}
	return url, nil
}
