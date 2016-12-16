package storage

import (
	"fmt"
	"github.com/minio/minio-go"
	"log"
	"strings"
)

// The S3 url protocol
const S3Protocol = "s3://"

// S3Backend provides access to an S3 object store.
type S3Backend struct {
	client *minio.Client
}

// NewS3Backend creates an S3Backend client instance, give an endpoint URL
// and a set of authentication credentials.
func NewS3Backend(endpoint string, id string, secret string, SSL bool) (*S3Backend, error) {

	// Initialize minio client object.
	client, err := minio.New(endpoint, id, secret, SSL)
	// TODO client needs to be closed?
	if err != nil {
		return nil, err
	}
	return &S3Backend{client}, nil
}

// Get copies an object from S3 to the host path.
func (s3 *S3Backend) Get(url string, hostPath string, class string) error {
	log.Printf("Starting download of %s", url)
	path := strings.TrimPrefix(url, S3Protocol)
	split := strings.SplitN(path, "/", 2)

	if class == File {
		err := s3.client.FGetObject(split[0], split[1], hostPath)
		if err != nil {
			return err
		}
		log.Printf("Successfully saved %s", hostPath)
		return nil
	} else if class == Directory {
		return fmt.Errorf("S3 directories not yet supported")
	}
	return fmt.Errorf("Unknown file class: %s", class)
}

// Put copies an object (file) from the host path to S3.
func (s3 *S3Backend) Put(url string, hostPath string, class string) error {
	log.Printf("Starting upload of %s", url)
	path := strings.TrimPrefix(url, S3Protocol)
  // TODO it's easy to create an error if this starts with a "/"
  //      maybe just strip it?
	split := strings.SplitN(path, "/", 2)

	if class == File {
		_, err := s3.client.FPutObject(split[0], split[1], hostPath, "application/data")
		if err != nil {
			return err
		}
		log.Printf("Successfully uploaded %s", hostPath)
		return nil
	} else if class == Directory {
		return fmt.Errorf("S3 directories not yet supported")
	}
	return fmt.Errorf("Unknown file class: %s", class)
}

func (s3 *S3Backend) Supports(url string, hostPath string, class string) bool {
	return strings.HasPrefix(url, S3Protocol)
}
