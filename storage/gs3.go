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

// GS3Protocol defines the expected URL prefix for S3, "s3://"
const GS3Protocol = "gs3://"

// S3Backend provides access to an S3 object store.
type GS3Backend struct {
	client *minio.Client
}

// NewGS3Backend creates an S3Backend client instance, give an endpoint URL
// and a set of authentication credentials.
func NewGS3Backend(conf config.GS3Storage) (*GS3Backend, error) {
	ssl := strings.HasPrefix(conf.Endpoint, "https")
	// Initialize minio client object.
	client, err := minio.NewV2(conf.Endpoint, conf.Key, conf.Secret, ssl)
	if err != nil {
		return nil, err
	}

	return &GS3Backend{client}, nil
}

// Get copies an object from S3 to the host path.
func (s3 *GS3Backend) Get(ctx context.Context, url string, hostPath string, class tes.FileType) error {

	path := strings.TrimPrefix(url, GS3Protocol)
	split := strings.SplitN(path, "/", 2)
	bucket := split[0]
	key := split[1]

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
func (s3 *GS3Backend) PutFile(ctx context.Context, url string, hostPath string) error {

	path := strings.TrimPrefix(url, GS3Protocol)
	split := strings.SplitN(path, "/", 2)
	bucket := split[0]
	key := split[1]

	_, err := s3.client.FPutObjectWithContext(ctx, bucket, key, hostPath, minio.PutObjectOptions{})

	return err
}

// Supports indicates whether this backend supports the given storage request.
// For S3, the url must start with "s3://".
func (s3 *GS3Backend) Supports(url string, hostPath string, class tes.FileType) bool {
	return strings.HasPrefix(url, GS3Protocol)
}
