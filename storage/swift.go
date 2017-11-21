package storage

import (
	"context"
	"fmt"
	"github.com/ncw/swift"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"io"
	"os"
	"path"
	"strings"
)

const swiftProtocol = "swift://"

// SwiftBackend provides access to an sw object store.
type SwiftBackend struct {
	conn *swift.Connection
}

// NewSwiftBackend creates an SwiftBackend client instance, give an endpoint URL
// and a set of authentication credentials.
func NewSwiftBackend(conf config.SwiftStorage) (*SwiftBackend, error) {

	// Create a connection
	c := swift.Connection{
		UserName: conf.UserName,
		ApiKey:   conf.Password,
		AuthUrl:  conf.AuthURL,
		Tenant:   conf.TenantName,
		TenantId: conf.TenantID,
		Region:   conf.RegionName,
	}

	// Read environment variables and apply them to the Connection structure.
	// Won't overwrite any parameters which are already set in the Connection struct.
	err := c.ApplyEnvironment()
	if err != nil {
		return nil, err
	}

	return &SwiftBackend{&c}, nil
}

// Get copies an object from storage to the host path.
func (sw *SwiftBackend) Get(ctx context.Context, rawurl string, hostPath string, class tes.FileType) error {
	if !sw.conn.Authenticated() {
		err := sw.conn.Authenticate()
		if err != nil {
			return fmt.Errorf("error connecting to Swift server: %v", err)
		}
	}

	url := sw.parse(rawurl)

	switch class {
	case tes.FileType_FILE:

		f, _, oerr := sw.conn.ObjectOpen(url.bucket, url.path, true, nil)
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
			f, _, oerr := sw.conn.ObjectOpen(url.bucket, obj.Name, true, nil)
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
	util.EnsurePath(hostPath)
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
	if !sw.conn.Authenticated() {
		err := sw.conn.Authenticate()
		if err != nil {
			return fmt.Errorf("error connecting to Swift server: %v", err)
		}
	}

	url := sw.parse(rawurl)

	reader, oerr := os.Open(hostPath)
	if oerr != nil {
		return oerr
	}

	writer, err := sw.conn.ObjectCreate(url.bucket, url.path, true, "", "", nil)
	if err != nil {
		return err
	}
	if _, cerr := io.Copy(writer, reader); cerr != nil {
		return cerr
	}
	if err := reader.Close(); err != nil {
		return err
	}
	return writer.Close()
}

// Supports indicates whether this backend supports the given storage request.
// For swift, the url must start with "swift://".
func (sw *SwiftBackend) Supports(rawurl string, hostPath string, class tes.FileType) bool {
	return strings.HasPrefix(rawurl, swiftProtocol)
}

func (sw *SwiftBackend) parse(rawurl string) *urlparts {
	path := strings.TrimPrefix(rawurl, swiftProtocol)
	split := strings.SplitN(path, "/", 2)
	bucket := split[0]
	key := split[1]
	return &urlparts{bucket, key}
}
