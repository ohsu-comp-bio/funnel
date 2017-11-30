package storage

import (
	"context"
	"fmt"
	"github.com/minio/minio-go"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"os"
	"path/filepath"
	"strings"
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
	url := s3.parse(rawurl)

	switch class {
	case File:
		err := s3.client.FGetObjectWithContext(ctx, url.bucket, url.path, hostPath, minio.GetObjectOptions{})
		if err != nil {
			return err
		}

	case Directory:
		// Create a done channel.
		doneCh := make(chan struct{})
		defer close(doneCh)
		// Recursively list all objects in 'mytestbucket'
		recursive := true
		for obj := range s3.client.ListObjects(url.bucket, url.path, recursive, doneCh) {
			// Create the directories in the path
			file := filepath.Join(hostPath, strings.TrimPrefix(obj.Key, url.path+"/"))
			if err := os.MkdirAll(filepath.Dir(file), 0775); err != nil {
				return err
			}
			err := s3.client.FGetObjectWithContext(ctx, url.bucket, obj.Key, file, minio.GetObjectOptions{})
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
	url := s3.parse(rawurl)
	_, err := s3.client.FPutObjectWithContext(ctx, url.bucket, url.path, hostPath, minio.PutObjectOptions{})

	return err
}

// SupportsGet indicates whether this backend supports GET storage request.
// For the GenericS3Backend, the url must start with "s3://" and the bucket must exist
func (s3 *GenericS3Backend) SupportsGet(rawurl string, class tes.FileType) error {
	if !strings.HasPrefix(rawurl, s3Protocol) {
		return fmt.Errorf("s3: unsupported protocol; expected %s", s3Protocol)
	}

	url := s3.parse(rawurl)
	ok, err := s3.client.BucketExists(url.bucket)
	if err != nil {
		return fmt.Errorf("s3: failed to find bucket: %s. error: %v", url.bucket, err)
	}
	if !ok {
		return fmt.Errorf("s3: bucket does not exist: %s", url.bucket)
	}

	return nil
}

// SupportsPut indicates whether this backend supports PUT storage request.
// For the GenericS3Backend, the url must start with "s3://" and the bucket must exist
func (s3 *GenericS3Backend) SupportsPut(rawurl string, class tes.FileType) error {
	return s3.SupportsGet(rawurl, class)
}

func (s3 *GenericS3Backend) parse(url string) *urlparts {
	path := strings.TrimPrefix(url, s3Protocol)
	path = strings.TrimPrefix(path, s3.endpoint)

	split := strings.SplitN(path, "/", 2)
	bucket := split[0]
	key := split[1]

	return &urlparts{bucket, key}
}
