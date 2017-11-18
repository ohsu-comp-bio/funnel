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

// NewGenericS3Backend creates an S3Backend client instance, give an endpoint URL
// and a set of authentication credentials.
func NewGenericS3Backend(conf config.S3Storage) (*GenericS3Backend, error) {
	ssl := strings.HasPrefix(conf.Endpoint, "https")

	client, err := minio.NewV2(conf.Endpoint, conf.Key, conf.Secret, ssl)
	if err != nil {
		return nil, fmt.Errorf("error creating s3 client: %v", err)
	}

	return &GenericS3Backend{client, conf.Endpoint}, nil
}

// Get copies an object from S3 to the host path.
func (s3 *GenericS3Backend) Get(ctx context.Context, url string, hostPath string, class tes.FileType) error {
	bucket, key := s3.processUrl(url)

	switch class {
	case File:
		err := s3.client.FGetObjectWithContext(ctx, bucket, key, hostPath, minio.GetObjectOptions{})
		if err != nil {
			return err
		}

	case Directory:
		// Create a done channel.
		doneCh := make(chan struct{})
		defer close(doneCh)
		// Recursively list all objects in 'mytestbucket'
		recursive := true
		for obj := range s3.client.ListObjectsV2(bucket, key, recursive, doneCh) {
			// Create the directories in the path
			file := filepath.Join(hostPath, strings.TrimPrefix(obj.Key, key+"/"))
			if err := os.MkdirAll(filepath.Dir(file), 0775); err != nil {
				return err
			}
			err := s3.client.FGetObjectWithContext(ctx, bucket, obj.Key, file, minio.GetObjectOptions{})
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}

	return nil
}

// Put copies an object (file) from the host path to S3.
func (s3 *GenericS3Backend) PutFile(ctx context.Context, url string, hostPath string) error {
	bucket, key := s3.processUrl(url)

	_, err := s3.client.FPutObjectWithContext(ctx, bucket, key, hostPath, minio.PutObjectOptions{})

	return err
}

// Supports indicates whether this backend supports the given storage request.
// For S3, the url must start with "s3://".
func (s3 *GenericS3Backend) Supports(url string, hostPath string, class tes.FileType) bool {
	if !strings.HasPrefix(url, S3Protocol) {
		return false
	}

	bucket, _ := s3.processUrl(url)
	found, _ := s3.client.BucketExists(bucket)

	return found
}

func (s3 *GenericS3Backend) processUrl(url string) (string, string) {
	path := strings.TrimPrefix(url, S3Protocol)
	path = strings.TrimPrefix(path, s3.endpoint+"/")

	split := strings.SplitN(path, "/", 2)
	bucket := split[0]
	key := split[1]

	return bucket, key
}
