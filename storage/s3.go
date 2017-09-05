package storage

import (
	"context"
	"fmt"
	"github.com/minio/minio-go"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"strings"
)

// S3Backend provides access to an S3 object store.
type S3Backend struct {
	client *minio.Client
}

// NewS3Backend creates an S3Backend client instance, give an endpoint URL
// and a set of authentication credentials.
func NewS3Backend(conf config.S3Storage) (*S3Backend, error) {

	// Initialize minio client object.
	// TODO SSL config and support
	client, err := minio.New(conf.Endpoint, conf.Key, conf.Secret, false)
	// TODO client needs to be closed?
	if err != nil {
		return nil, err
	}
	return &S3Backend{client}, nil
}

// Get copies an object from S3 to the host path.
func (s3 *S3Backend) Get(ctx context.Context, url string, hostPath string, class tes.FileType) error {
	log.Info("Starting download", "url", url)
	path := strings.TrimPrefix(url, "s3://")
	split := strings.SplitN(path, "/", 2)

	if class == file {
		err := s3.client.FGetObject(split[0], split[1], hostPath)
		if err != nil {
			return err
		}
		log.Info("Successfully saved", "hostPath", hostPath)
		return nil
	} else if class == directory {
		return fmt.Errorf("S3 directories not yet supported")
	}
	return fmt.Errorf("Unknown file class: %s", class)
}

// Put copies an object (file) from the host path to S3.
func (s3 *S3Backend) Put(ctx context.Context, url string, hostPath string, class tes.FileType) ([]*tes.OutputFileLog, error) {

	log.Info("Starting upload", "url", url)
	path := strings.TrimPrefix(url, "s3://")
	// TODO it's easy to create an error if this starts with a "/"
	//      maybe just strip it?
	split := strings.SplitN(path, "/", 2)

	switch class {
	case file:
		_, err := s3.client.FPutObject(split[0], split[1], hostPath, "application/data")
		if err != nil {
			return nil, err
		}
		log.Info("Successfully uploaded", "hostPath", hostPath)
		return []*tes.OutputFileLog{
			{Url: url, Path: hostPath, SizeBytes: fileSize(hostPath)},
		}, nil

	case directory:
		return nil, fmt.Errorf("S3 directories not yet supported")

	default:
		return nil, fmt.Errorf("Unknown file class: %s", class)
	}
}

// Supports indicates whether this backend supports the given storage request.
// For S3, the url must start with "s3://".
func (s3 *S3Backend) Supports(url string, hostPath string, class tes.FileType) bool {
	return strings.HasPrefix(url, "s3://")
}
