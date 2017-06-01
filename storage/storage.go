package storage

// NOTE!
// It's important that Storage instances be immutable!
// We don't want storage authentication to be accidentally shared between tasks.
// If they are mutable, there's more chance that storage config can leak
// between separate processes.

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"os"
	"path/filepath"
)

const (
	// File represents the file type
	File = tes.FileType_FILE
	// Directory represents the directory type
	Directory = tes.FileType_DIRECTORY
)

// Backend provides an interface for a storage backend.
// New storage backends must support this interface.
type Backend interface {
	Get(ctx context.Context, url string, path string, class tes.FileType) error
	Put(ctx context.Context, url string, path string, class tes.FileType) ([]*tes.OutputFileLog, error)
	// Determines whether this backends supports the given request (url/path/class).
	// A backend normally uses this to match the url prefix (e.g. "s3://")
	// TODO would it be useful if this included the request type (Get/Put)?
	Supports(url string, path string, class tes.FileType) bool
}

// Storage provides a client for accessing multiple storage systems,
// i.e. for downloading/uploading task files from S3, GS, local disk, etc.
//
// For a given storage url, the storage backend is usually determined by the url prefix,
// e.g. "s3://my-bucket/file" will access the S3 backend.
type Storage struct {
	backends []Backend
}

// Get downloads a file from a storage system at the given "url".
// The file is downloaded to the given local "path".
// "class" is either "File" or "Directory".
func (storage Storage) Get(ctx context.Context, url string, path string, class tes.FileType) error {
	backend, err := storage.findBackend(url, path, class)
	if err != nil {
		return err
	}
	return backend.Get(ctx, url, path, class)
}

// Put uploads a file to a storage system at the given "url".
// The file is uploaded from the given local "path".
// "class" is either "File" or "Directory".
func (storage Storage) Put(ctx context.Context, url string, path string, class tes.FileType) ([]*tes.OutputFileLog, error) {
	backend, err := storage.findBackend(url, path, class)
	if err != nil {
		return nil, err
	}
	return backend.Put(ctx, url, path, class)
}

// Supports indicates whether the storage supports the given request.
func (storage Storage) Supports(url string, path string, class tes.FileType) bool {
	b, _ := storage.findBackend(url, path, class)
	return b != nil
}

// findBackend tries to find a backend that matches the given url/path/class.
// This is how a url gets matched to a backend, for example by the url prefix "s3://".
func (storage Storage) findBackend(url string, path string, class tes.FileType) (Backend, error) {
	for _, backend := range storage.backends {
		if backend.Supports(url, path, class) {
			return backend, nil
		}
	}
	return nil, fmt.Errorf("Could not find matching storage system for %s", url)
}

// WithBackend returns a new child Storage instance with the given backend added.
func (storage Storage) WithBackend(b Backend) Storage {
	storage.backends = append(storage.backends, b)
	return storage
}

// WithConfig returns a new Storage instance with the given additional configuration.
func (storage Storage) WithConfig(conf config.StorageConfig) (Storage, error) {

	if conf.Local.Valid() {
		local, err := NewLocalBackend(conf.Local)
		if err != nil {
			return storage, err
		}
		storage = storage.WithBackend(local)
	}

	for _, c := range conf.S3 {
		if c.Valid() {
			s3, err := NewS3Backend(c)
			if err != nil {
				return storage, err
			}
			storage = storage.WithBackend(s3)
		}
	}

	for _, c := range conf.GS {
		if c.Valid() {
			gs, nerr := NewGSBackend(c)
			if nerr != nil {
				return storage, nerr
			}
			storage = storage.WithBackend(gs)
		}
	}

	return storage, nil
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
