package storage

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/tes"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

// GenericS3Backend provides access to an S3 object store.
type GenericS3Backend struct {
	client   *minio.Client
	endpoint string
}

// NewGenericS3Backend creates a new GenericS3Backend instance, given an endpoint URL
// and a set of authentication credentials.
func NewGenericS3Backend(conf config.GenericS3Storage) (*GenericS3Backend, error) {
	ssl := strings.HasPrefix(conf.Endpoint, "https")
	endpoint := endpointRE.ReplaceAllString(conf.Endpoint, "$2")
	client, err := minio.NewV2(endpoint, conf.Key, conf.Secret, ssl)
	if err != nil {
		return nil, fmt.Errorf("error creating generic s3 backend: %v", err)
	}

	return &GenericS3Backend{client, endpoint + "/"}, nil
}

// Get copies an object from S3 to the host path.
func (s3 *GenericS3Backend) Get(ctx context.Context, rawurl string, hostPath string, class tes.FileType) error {
	url, err := s3.parse(rawurl)
	if err != nil {
		return err
	}

	switch class {
	case File:
		err := fsutil.EnsurePath(hostPath)
		if err != nil {
			return err
		}
		return s3.client.FGetObjectWithContext(ctx, url.bucket, url.path, hostPath, minio.GetObjectOptions{})

	case Directory:
		err := fsutil.EnsureDir(hostPath)
		if err != nil {
			return err
		}
		// Create a done channel.
		doneCh := make(chan struct{})
		defer close(doneCh)
		// Recursively list all objects in 'mytestbucket'
		recursive := true
		objects := []minio.ObjectInfo{}
		for obj := range s3.client.ListObjects(url.bucket, url.path, recursive, doneCh) {
			objects = append(objects, obj)
		}

		if len(objects) == 0 {
			return ErrEmptyDirectory
		}

		for _, obj := range objects {
			// Create the directories in the path
			file := filepath.Join(hostPath, strings.TrimPrefix(obj.Key, url.path+"/"))
			// check if key represents a directory
			if strings.HasSuffix(obj.Key, "/") {
				continue
			}
			err = fsutil.EnsurePath(file)
			if err != nil {
				return err
			}

			err = s3.client.FGetObjectWithContext(ctx, url.bucket, obj.Key, file, minio.GetObjectOptions{})
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}

	return nil
}

// PutFile copies an object (file) from the host path to S3.
func (s3 *GenericS3Backend) PutFile(ctx context.Context, rawurl string, hostPath string) error {
	url, err := s3.parse(rawurl)
	if err != nil {
		return err
	}

	_, err = s3.client.FPutObjectWithContext(ctx, url.bucket, url.path, hostPath, minio.PutObjectOptions{})
	return err
}

// SupportsGet indicates whether this backend supports GET storage request.
// For the GenericS3Backend, the url must start with "s3://" and the bucket must exist
func (s3 *GenericS3Backend) SupportsGet(rawurl string, class tes.FileType) error {
	url, err := s3.parse(rawurl)
	if err != nil {
		return err
	}
	ok, err := s3.client.BucketExists(url.bucket)
	if err != nil {
		return fmt.Errorf("genericS3: failed to find bucket: %s. error: %v", url.bucket, err)
	}
	if !ok {
		return fmt.Errorf("genericS3: bucket does not exist: %s", url.bucket)
	}
	return nil
}

// SupportsPut indicates whether this backend supports PUT storage request.
// For the GenericS3Backend, the url must start with "s3://" and the bucket must exist
func (s3 *GenericS3Backend) SupportsPut(rawurl string, class tes.FileType) error {
	return s3.SupportsGet(rawurl, class)
}

func (s3 *GenericS3Backend) parse(rawurl string) (*urlparts, error) {
	if !strings.HasPrefix(rawurl, s3Protocol) {
		return nil, &ErrUnsupportedProtocol{"genericS3"}
	}

	path := strings.TrimPrefix(rawurl, s3Protocol)
	path = strings.TrimPrefix(path, s3.endpoint)
	if path == "" {
		return nil, &ErrInvalidURL{"genericS3"}
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
