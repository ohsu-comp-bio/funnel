package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/alecthomas/units"
	"github.com/ncw/swift"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

const swiftProtocol = "swift://"

// SwiftBackend provides access to an sw object store.
type SwiftBackend struct {
	conn      *swift.Connection
	chunkSize int64
}

// NewSwiftBackend creates an SwiftBackend client instance, give an endpoint URL
// and a set of authentication credentials.
func NewSwiftBackend(conf config.SwiftStorage) (Backend, error) {

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
		chunkSize = int64(100 * units.MB)
	} else if conf.ChunkSizeBytes > int64(5*units.GB) {
		chunkSize = int64(5 * units.GB)
	} else {
		chunkSize = conf.ChunkSizeBytes
	}

	b := &SwiftBackend{conn, chunkSize}
	return &retrier{
		backend: b,
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

// Get copies an object from storage to the host path.
func (sw *SwiftBackend) Get(ctx context.Context, rawurl string, hostPath string, class tes.FileType) error {
	url, err := sw.parse(rawurl)
	if err != nil {
		return err
	}

	var checkHash = true
	var headers swift.Headers

	switch class {
	case File:
		f, _, err := sw.conn.ObjectOpen(url.bucket, url.path, checkHash, headers)
		if err != nil {
			return err
		}
		defer f.Close()

		return sw.get(ctx, f, hostPath)

	case Directory:
		err := fsutil.EnsureDir(hostPath)
		if err != nil {
			return err
		}

		objs, err := sw.conn.ObjectsAll(url.bucket, &swift.ObjectsOpts{
			Prefix: url.path,
		})
		if err != nil {
			return fmt.Errorf("listing objects by prefix: %s", err)
		}
		if len(objs) == 0 {
			return ErrEmptyDirectory
		}

		for _, obj := range objs {
			if strings.HasSuffix(obj.Name, "/") {
				continue
			}
			f, _, err := sw.conn.ObjectOpen(url.bucket, obj.Name, checkHash, headers)
			if err != nil {
				return fmt.Errorf("opening object %s: %s", obj.Name, err)
			}

			if err = sw.get(ctx, f, path.Join(hostPath, strings.TrimPrefix(obj.Name, url.path))); err != nil {
				if cerr := f.Close(); cerr != nil {
					return fmt.Errorf("closing object after get %v; %v", err, cerr)
				}
				return err
			}

			if err = f.Close(); err != nil {
				return fmt.Errorf("closing file: %s", err)
			}
		}

		return nil

	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}
}

func (sw *SwiftBackend) get(ctx context.Context, src io.Reader, hostPath string) (err error) {
	err = fsutil.EnsurePath(hostPath)
	if err != nil {
		return err
	}
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

	_, err = io.Copy(dest, fsutil.Reader(ctx, src))
	return err
}

// PutFile copies an object (file) from the host path to storage.
func (sw *SwiftBackend) PutFile(ctx context.Context, rawurl string, hostPath string) error {
	url, err := sw.parse(rawurl)
	if err != nil {
		return err
	}

	reader, err := os.Open(hostPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	var writer io.WriteCloser
	var checkHash = true
	var hash string
	var contentType string
	var headers swift.Headers

	fSize := fileSize(hostPath)
	if fSize < int64(5*units.GB) {
		writer, err = sw.conn.ObjectCreate(url.bucket, url.path, checkHash, hash, contentType, headers)
	} else {
		writer, err = sw.conn.StaticLargeObjectCreateFile(&swift.LargeObjectOpts{
			Container:  url.bucket,
			ObjectName: url.path,
			CheckHash:  checkHash,
			Hash:       hash,
			Headers:    headers,
			ChunkSize:  sw.chunkSize,
		})
	}
	if err != nil {
		return err
	}
	defer func() {
		cerr := writer.Close()
		if cerr != nil {
			err = fmt.Errorf("%v; %v", err, cerr)
		}
	}()

	_, err = io.Copy(writer, fsutil.Reader(ctx, reader))
	return err
}

// SupportsGet indicates whether this backend supports GET storage request.
// For the Swift backend, the url must start with "swift://" and the bucket must exist
func (sw *SwiftBackend) SupportsGet(rawurl string, class tes.FileType) error {
	url, err := sw.parse(rawurl)
	if err != nil {
		return err
	}
	_, _, err = sw.conn.Container(url.bucket)
	if err != nil {
		return fmt.Errorf("swift: failed to find bucket: %s. error: %v", url.bucket, err)
	}
	return nil
}

// SupportsPut indicates whether this backend supports PUT storage request.
// For the Swift backend, the url must start with "swift://" and the bucket must exist
func (sw *SwiftBackend) SupportsPut(rawurl string, class tes.FileType) error {
	return sw.SupportsGet(rawurl, class)
}

func (sw *SwiftBackend) parse(rawurl string) (*urlparts, error) {
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
