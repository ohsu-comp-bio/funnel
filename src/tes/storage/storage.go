package storage

import (
	"context"
	"fmt"
	"tes/config"
)

// NOTE!
// It's important that Storage instances be immutable!
// We dont want storage authentication to be accidentally shared between jobs.
// If they are mutable, there's more chance that storage config can leak
// between separate processes.

const (
	// File represents the file type
	File string = "File"
	// ReadOnlyFile represents a file in a read only volume
	ReadOnlyFile string = "ReadOnlyFile"
	// Directory represents the directory type
	Directory = "Directory"
)

// Backend provides an interface for a storage backend.
// New storage backends must support this interface.
type Backend interface {
	Get(ctx context.Context, url string, path string, class string) error
	Put(ctx context.Context, url string, path string, class string) error
	// Determines whether this backends supports the given request (url/path/class).
	// A backend normally uses this to match the url prefix (e.g. "s3://")
	// TODO would it be useful if this included the request type (Get/Put)?
	Supports(url string, path string, class string) bool
}

// Storage provides a client for accessing multiple storage systems,
// i.e. for downloading/uploading job files from multiple S3, NFS, local disk, etc.
//
// For a given storage url, the storage backend is usually determined by the url prefix,
// e.g. "s3://my-bucket/file" will access the S3 backend.
type Storage struct {
	backends []Backend
}

// Get downloads a file from a storage system at the given "url".
// The file is downloaded to the given local "path".
// "class" is either "File", "ReadOnlyFile" or "Directory".
func (storage Storage) Get(ctx context.Context, url string, path string, class string) error {
	backend, err := storage.findBackend(url, path, class)
	if err != nil {
		return err
	}
	return backend.Get(ctx, url, path, class)
}

// Put uploads a file to a storage system at the given "url".
// The file is uploaded from the given local "path".
// "class" is either "File" or "Directory".
func (storage Storage) Put(ctx context.Context, url string, path string, class string) error {
	backend, err := storage.findBackend(url, path, class)
	if err != nil {
		return err
	}
	return backend.Put(ctx, url, path, class)
}

// Supports indicates whether the storage supports the given request.
func (storage Storage) Supports(url string, path string, class string) bool {
	b, _ := storage.findBackend(url, path, class)
	return b != nil
}

// findBackend tries to find a backend that matches the given url/path/class.
// This is how a url gets matched to a backend, for example by the url prefix "s3://".
func (storage Storage) findBackend(url string, path string, class string) (Backend, error) {
	for _, backend := range storage.backends {
		if backend.Supports(url, path, class) {
			return backend, nil
		}
	}
	return nil, fmt.Errorf("Could not find matching storage system for %s", url)
}

// WithBackend returns a new child Storage instance with the given backend added.
func (storage Storage) WithBackend(b Backend) (*Storage, error) {
	backends := append(storage.backends, b)
	return &Storage{backends}, nil
}

// WithConfig returns a new Storage instance with the given additional configuration.
func (storage Storage) WithConfig(conf *config.StorageConfig) (*Storage, error) {
	var err error
	var out *Storage

	if conf.Local.Valid() {
		local, err := NewLocalBackend(conf.Local)
		if err != nil {
			return nil, err
		}
		out, err = storage.WithBackend(local)
	}

	if conf.S3.Valid() {
		s3, err := NewS3Backend(conf.S3)
		if err != nil {
			return nil, err
		}
		out, err = storage.WithBackend(s3)
	}

	if conf.GS.Valid() {
		gs, nerr := NewGSBackend(conf.GS)
		if nerr != nil {
			return nil, nerr
		}
		out, err = storage.WithBackend(gs)
	}

	if err != nil {
		return nil, err
	}

	// If the configuration did nothing, return the initial storage instance
	if out == nil {
		return &storage, nil
	}

	return out, nil
}
