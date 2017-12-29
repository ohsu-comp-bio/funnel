package storage

// NOTE!
// It's important that Storage instances be immutable!
// We don't want storage authentication to be accidentally shared between tasks.
// If they are mutable, there's more chance that storage config can leak
// between separate processes.

import (
	"context"
	"errors"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"os"
	"path/filepath"
	"strings"
)

const (
	// File represents the file type
	File = tes.FileType_FILE
	// Directory represents the directory type
	Directory = tes.FileType_DIRECTORY
)

// ErrEmptyDirectory is returned by Get/Put in these cases:
//   1. Local file system - Directory exists, but is empty
//   2. Object store - No objects match prefix
var ErrEmptyDirectory = errors.New("no files found in directory")

// Backend provides an interface for a storage backend.
// New storage backends must support this interface.
type Backend interface {
	Get(ctx context.Context, url string, path string, class tes.FileType) error
	PutFile(ctx context.Context, url string, path string) error
	// Determines whether this backends supports the given request (url/path/class).
	// A backend normally uses this to match the url prefix (e.g. "s3://")
	SupportsGet(url string, class tes.FileType) error
	SupportsPut(url string, class tes.FileType) error
}

// Storage provides a client for accessing multiple storage systems,
// i.e. for downloading/uploading task files from S3, GS, local disk, etc.
//
// For a given storage url, the storage backend is usually determined by the url prefix,
// e.g. "s3://my-bucket/file" will access the S3 backend.
type Storage struct {
	backends []Backend
}

// NewStorage returns a new Storage instance with the given additional configuration.
func NewStorage(conf config.Config) (Storage, error) {
	storage := Storage{}

	if conf.LocalStorage.Valid() {
		local, err := NewLocalBackend(conf.LocalStorage)
		if err != nil {
			return storage, fmt.Errorf("failed to configure local storage backend: %s", err)
		}
		storage = storage.WithBackend(local)
	}

	if conf.AmazonS3.Valid() {
		s3, err := NewAmazonS3Backend(conf.AmazonS3)
		if err != nil {
			return storage, fmt.Errorf("failed to configure Amazon S3 storage backend: %s", err)
		}
		storage = storage.WithBackend(s3)
	}

	if conf.GoogleStorage.Valid() {
		gs, nerr := NewGSBackend(conf.GoogleStorage)
		if nerr != nil {
			return storage, fmt.Errorf("failed to configure Google Storage backend: %s", nerr)
		}
		storage = storage.WithBackend(gs)
	}

	if conf.Swift.Valid() {
		s, err := NewSwiftBackend(conf.Swift)
		if err != nil {
			return storage, fmt.Errorf("failed to config Swift storage backend: %s", err)
		}
		storage = storage.WithBackend(s)
	}

	for _, c := range conf.GenericS3 {
		if c.Valid() {
			s, err := NewGenericS3Backend(c)
			if err != nil {
				return storage, fmt.Errorf("failed to config generic S3 storage backend: %s", err)
			}
			storage = storage.WithBackend(s)
		}
	}

	if conf.HTTPStorage.Valid() {
		http, err := NewHTTPBackend(conf.HTTPStorage)
		if err != nil {
			return storage, fmt.Errorf("failed to config http storage backend: %s", err)
		}
		storage = storage.WithBackend(http)
	}

	return storage, nil
}

// WithBackend returns a new child Storage instance with the given backend added.
func (storage Storage) WithBackend(b Backend) Storage {
	storage.backends = append(storage.backends, b)
	return storage
}

// Get downloads a file from a storage system at the given "url".
// The file is downloaded to the given local "path".
// "class" is either "File" or "Directory".
func (storage Storage) Get(ctx context.Context, url string, path string, class tes.FileType) error {
	backend, err := storage.findBackend(url, class, "get")
	if err != nil {
		return err
	}

	return backend.Get(ctx, url, path, class)
}

// Put uploads a file to a storage system at the given "url".
// The file is uploaded from the given local "path".
// "class" is either "File" or "Directory".
func (storage Storage) Put(ctx context.Context, url string, path string, class tes.FileType) ([]*tes.OutputFileLog, error) {
	backend, err := storage.findBackend(url, class, "put")
	if err != nil {
		return nil, err
	}

	var out []*tes.OutputFileLog

	switch class {
	case File:
		err = backend.PutFile(ctx, url, path)
		if err != nil {
			return nil, err
		}
		out = append(out, &tes.OutputFileLog{
			Url:       url,
			Path:      path,
			SizeBytes: fileSize(path),
		})

	case Directory:
		var files []hostfile
		files, err = walkFiles(path)
		if err != nil {
			return nil, err
		}

		if len(files) == 0 {
			return nil, ErrEmptyDirectory
		}

		for _, f := range files {
			u := strings.TrimSuffix(url, "/") + "/" + f.rel
			err = backend.PutFile(ctx, u, f.abs)
			if err != nil {
				return nil, err
			}

			out = append(out, &tes.OutputFileLog{
				Url:       u,
				Path:      f.abs,
				SizeBytes: f.size,
			})
		}

	default:
		return nil, fmt.Errorf("Unknown file class: %s", class)
	}

	return out, nil
}

// SupportsGet indicates whether the storage supports the given request.
func (storage Storage) SupportsGet(url string, class tes.FileType) error {
	_, err := storage.findBackend(url, class, "get")
	return err
}

// SupportsPut indicates whether the storage supports the given request.
func (storage Storage) SupportsPut(url string, class tes.FileType) error {
	_, err := storage.findBackend(url, class, "put")
	return err
}

func (storage Storage) findBackend(url string, class tes.FileType, op string) (Backend, error) {
	var found = 0
	var useBackend Backend
	var err error
	var errs []string

	for _, backend := range storage.backends {
		switch op {
		case "get":
			err = backend.SupportsGet(url, class)
		case "put":
			err = backend.SupportsPut(url, class)
		default:
			return nil, fmt.Errorf("unknown operation: %s; expected 'get' or 'put'", op)
		}
		if err == nil {
			useBackend = backend
			found++
		} else {
			errs = append(errs, err.Error())
		}
	}

	if found == 0 {
		return nil, fmt.Errorf("Could not find matching storage system for: %s\n%s", url, strings.Join(errs, "\n"))
	} else if found > 1 {
		return nil, fmt.Errorf("Request supported by multiple backends for: %s", url)
	}

	return useBackend, nil
}

type hostfile struct {
	// The path relative to the "root" given to walkFiles().
	rel string
	// The absolute path of the file on the host.
	abs string
	// Size of the file in bytes
	size int64
}

func walkFiles(root string) ([]hostfile, error) {
	var files []hostfile

	if dinfo, err := os.Stat(root); os.IsNotExist(err) || !dinfo.IsDir() {
		return nil, fmt.Errorf("%s does not exist or is not a directory", root)
	}

	err := filepath.Walk(root, func(p string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			rel, err := filepath.Rel(root, p)
			if err != nil {
				return err
			}
			files = append(files, hostfile{rel, p, f.Size()})
		}
		return nil
	})
	return files, err
}

// Get the file size, or return 0 if there's an error calling os.Stat().
func fileSize(path string) int64 {
	st, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return st.Size()
}

type urlparts struct {
	bucket string
	path   string
}
