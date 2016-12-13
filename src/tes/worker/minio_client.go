package tesTaskEngineWorker

import (
	"fmt"
	"github.com/minio/minio-go"
	"log"
	"strings"
)

// S3Protocol documentation
// TODO: documentation
var S3Protocol = "s3://"

// S3Access documentation
// TODO: documentation
type S3Access struct {
	client *minio.Client
}

// NewS3Access documentation
// TODO: documentation
func NewS3Access(endpoint string, accessKeyID string, secretAccessKey string, useSSL bool) *S3Access {

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, accessKeyID, secretAccessKey, useSSL)
	if err != nil {
		log.Fatalln(err)
	}
	return &S3Access{client: minioClient}
}

// Get documentation
// TODO: documentation
func (s3access *S3Access) Get(storage string, hostPath string, class string) error {
	log.Printf("Starting download of %s", storage)
	storage = strings.TrimPrefix(storage, SwiftProtocol)
	storageSplit := strings.SplitN(storage, "/", 2)

	if class == "File" {
		// Downloads everything into a DownloadResult struct.
		if err := s3access.client.FGetObject(storageSplit[0], storageSplit[1], hostPath); err != nil {
			return err
		}
		log.Println("Successfully saved %s", hostPath)
		return nil
	} else if class == "Directory" {
		return fmt.Errorf("S3 directories not yet supported")
	}
	return fmt.Errorf("Unknown element type: %s", class)
}

// Put documentation
// TODO: documentation
func (s3access *S3Access) Put(storage string, hostPath string, class string) error {
	log.Printf("Starting upload of %s", storage)
	storage = strings.TrimPrefix(storage, SwiftProtocol)
	storageSplit := strings.SplitN(storage, "/", 2)
	if class == "File" {
		if _, err := s3access.client.FPutObject(storageSplit[0], storageSplit[1], hostPath, "application/data"); err != nil {
			return err
		}
		log.Println("Successfully uploaded %s", hostPath)
		return nil
	} else if class == "Directory" {
		return fmt.Errorf("S3 directories not yet supported")
	}
	return fmt.Errorf("Unknown element type: %s", class)
}
