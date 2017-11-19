package storage

import (
	"context"
	"fmt"
	"github.com/ncw/swift"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
	"io"
	"os"
	"path"
	"strings"
)

const swiftProtocol = "swift://"

// SwiftBackend provides access to an sw object store.
type SwiftBackend struct {
	conn      *swift.Connection
	chunkSize int64
}

// NewSwiftBackend creates an SwiftBackend client instance, give an endpoint URL
// and a set of authentication credentials.
func NewSwiftBackend(conf config.SwiftStorage) (*SwiftBackend, error) {

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
	if conf.ChunkSizeBytes < 10000000 {
		chunkSize = 500000000 // 500 MB
	} else {
		chunkSize = conf.ChunkSizeBytes
	}

	return &SwiftBackend{conn, chunkSize}, nil
}

// Get copies an object from storage to the host path.
func (sw *SwiftBackend) Get(ctx context.Context, rawurl string, hostPath string, class tes.FileType) error {
	url := sw.parse(rawurl)

	var checkHash = true
	var headers swift.Headers

	switch class {
	case tes.FileType_FILE:

		f, _, oerr := sw.conn.ObjectOpen(url.bucket, url.path, checkHash, headers)
		if oerr != nil {
			return oerr
		}

		if err := sw.get(f, hostPath); err != nil {
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}

		return nil

	case tes.FileType_DIRECTORY:
		objs, err := sw.conn.ObjectsAll(url.bucket, &swift.ObjectsOpts{
			Prefix: url.path,
		})
		if err != nil {
			return err
		}

		for _, obj := range objs {
			f, _, oerr := sw.conn.ObjectOpen(url.bucket, obj.Name, checkHash, headers)
			if oerr != nil {
				return oerr
			}

			if err := sw.get(f, path.Join(hostPath, strings.TrimPrefix(obj.Name, url.path))); err != nil {
				return err
			}

			if err := f.Close(); err != nil {
				return err
			}
		}

		return nil

	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}
}

func (sw *SwiftBackend) get(src io.Reader, hostPath string) error {
	fsutil.EnsurePath(hostPath)
	dest, cerr := os.Create(hostPath)
	if cerr != nil {
		return cerr
	}

	_, werr := io.Copy(dest, src)
	if werr != nil {
		return werr
	}
	return dest.Close()
}

// PutFile copies an object (file) from the host path to storage.
func (sw *SwiftBackend) PutFile(ctx context.Context, rawurl string, hostPath string) error {
	url := sw.parse(rawurl)

	reader, oerr := os.Open(hostPath)
	if oerr != nil {
		return oerr
	}

	var writer io.WriteCloser
	var err error

	var checkHash = true
	var hash string
	var contentType string
	var headers swift.Headers

	fSize := fileSize(hostPath)
	// upload as chunks if file size is > 4.95 GB (swift has a 5 GB limit for objects)
	if fSize < 4950000000 {
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

	if _, cerr := io.Copy(writer, reader); cerr != nil {
		return cerr
	}
	if rerr := reader.Close(); rerr != nil {
		return rerr
	}
	return writer.Close()
}

// Supports returns true if this backend supports the given storage request.
// For the Swift backend, the url must start with "swift://"
func (sw *SwiftBackend) Supports(rawurl string) error {
	ok := strings.HasPrefix(rawurl, swiftProtocol)
	if !ok {
		return fmt.Errorf("swift: unsupported protocol; expected %s", swiftProtocol)
	}
	url := sw.parse(rawurl)
	_, _, err := sw.conn.Container(url.bucket)
	if err != nil {
		return fmt.Errorf("swift: failed to find bucket: %s. error: %v", url.bucket, err)
	}
	return nil
}

func (sw *SwiftBackend) parse(rawurl string) *urlparts {
	path := strings.TrimPrefix(rawurl, swiftProtocol)
	split := strings.SplitN(path, "/", 2)
	bucket := split[0]
	key := split[1]
	return &urlparts{bucket, key}
}
