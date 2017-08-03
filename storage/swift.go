package storage

import (
	"context"
	"fmt"
	"github.com/ncw/swift"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"io"
	urllib "net/url"
	"os"
	"path"
	"strings"
)

const swiftScheme = "swift"

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

	// Authenticate
	err := c.Authenticate()
	if err != nil {
		return nil, err
	}
	return &SwiftBackend{&c}, nil
}

// Get copies an object from storage to the host path.
func (sw *SwiftBackend) Get(ctx context.Context, rawurl string, hostPath string, class tes.FileType) error {

	url, perr := sw.parse(rawurl)
	if perr != nil {
		return perr
	}

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

// Put copies an object (file) from the host path to storage.
func (sw *SwiftBackend) Put(ctx context.Context, rawurl string, hostPath string, class tes.FileType) ([]*tes.OutputFileLog, error) {

	var out []*tes.OutputFileLog

	switch class {
	case tes.FileType_FILE:

		err := sw.put(rawurl, hostPath)
		if err != nil {
			return nil, err
		}
		out = append(out, &tes.OutputFileLog{
			Url:       rawurl,
			Path:      hostPath,
			SizeBytes: fileSize(hostPath),
		})

	case tes.FileType_DIRECTORY:
		files, err := walkFiles(hostPath)

		for _, f := range files {
			u := rawurl + "/" + f.rel
			out = append(out, &tes.OutputFileLog{
				Url:       u,
				Path:      f.abs,
				SizeBytes: f.size,
			})
			err := sw.put(u, f.abs)
			if err != nil {
				return nil, err
			}
		}

		if err != nil {
			return nil, err
		}

	default:
		return nil, fmt.Errorf("Unknown file class: %s", class)
	}

	return out, nil
}

func (sw *SwiftBackend) put(rawurl, hostPath string) error {

	url, perr := sw.parse(rawurl)
	if perr != nil {
		return perr
	}

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

func (sw *SwiftBackend) parse(rawurl string) (*urlparts, error) {
	url, err := urllib.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	if url.Scheme != swiftScheme {
		return nil, fmt.Errorf("Invalid URL scheme '%s' for Swift Storage backend in url: %s", url.Scheme, rawurl)
	}

	bucket := url.Host
	path := strings.TrimLeft(url.EscapedPath(), "/")
	return &urlparts{bucket, path}, nil
}

// Supports indicates whether this backend supports the given storage request.
// For sw, the url must start with "sw://".
func (sw *SwiftBackend) Supports(rawurl string, hostPath string, class tes.FileType) bool {
	_, err := sw.parse(rawurl)
	if err != nil {
		return false
	}
	return true
}
