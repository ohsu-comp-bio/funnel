package storage

import (
	"fmt"
	pbr "tes/server/proto"
)

const (
	File      string = "File"
	Directory        = "Directory"
)

// Backend provides an interface for a storage backend.
// New storage backends must support this interface.
type Backend interface {
	Get(url string, path string, class string) error
	Put(url string, path string, class string) error
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
	stores []Backend
}

// Get downloads a file from a storage system at the given "url".
// The file is downloaded to the given local "path".
// "class" is either "File" or "Directory".
func (storage Storage) Get(url string, path string, class string) error {
	store, err := storage.findBackend(url, path, class)
	if err != nil {
		return err
	}
	return store.Get(url, path, class)
}

// Put uploads a file to a storage system at the given "url".
// The file is uploaded from the given local "path".
// "class" is either "File" or "Directory".
func (storage Storage) Put(url string, path string, class string) error {
	store, err := storage.findBackend(url, path, class)
	if err != nil {
		return err
	}
	return store.Put(url, path, class)
}

// findBackend tries to find a backend that matches the given url/path/class.
// This is how a url gets matched to a backend, for example by the url prefix "s3://".
func (storage Storage) findBackend(url string, path string, class string) (Backend, error) {
	for _, store := range storage.stores {
		if store.Supports(url, path, class) {
			return store, nil
		}
	}
	return nil, fmt.Errorf("Could not find matching storage system for %s", url)
}

// Backend config functions below
// It's important that these create a new Storage instance, because we don't
// want storage authentication to be accidentally shared between jobs, so
// a job-specific Storage instance should be configured and destroyed for each job.

// WithS3 returns a new child Storage instance with the given
// S3 backend config added.
func (storage Storage) WithS3(en string, id string, sc string, ssl bool) (*Storage, error) {
	s3, err := NewS3Backend(en, id, sc, ssl)
	if err != nil {
		return nil, err
	}
	stores := append(storage.stores, s3)
	return &Storage{stores}, nil
}

// WithLocal returns a new child Storage instance with the given local backend config added.
func (storage Storage) WithLocal(allow []string) (*Storage, error) {
	local := NewLocalBackend(allow)
	stores := append(storage.stores, local)
	return &Storage{stores}, nil
}

func (storage Storage) WithConfig(conf *pbr.StorageConfig) (*Storage, error) {
	var err error
	var out *Storage

	switch x := conf.Protocol.(type) {
	case *pbr.StorageConfig_S3:
		// TODO need config validation to ensure these values actually exist
		out, err = storage.WithS3(
			x.S3.Endpoint,
			x.S3.Key,
			x.S3.Secret,
			false, // use SSL?
		)
	case *pbr.StorageConfig_Local:
		out, err = storage.WithLocal(x.Local.AllowedDirs)
	case nil:
	default:
		err = fmt.Errorf("Unknown storage protocol in config: %T", x)
	}
	if err != nil {
		return nil, err
	}
	return out, nil
}
