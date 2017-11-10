package storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
	"os"
	"path/filepath"
	"strings"
)

// S3Protocol defines the expected URL prefix for S3, "s3://"
const S3Protocol = "s3://"

// S3Backend provides access to an S3 object store.
type S3Backend struct {
	sess *session.Session
}

// NewS3Backend creates an S3Backend session instance
func NewS3Backend(conf config.S3Storage) (*S3Backend, error) {
	awsConf := util.NewAWSConfigWithCreds(conf.Credentials.Key, conf.Credentials.Secret)
	awsConf.WithMaxRetries(10)

	sess, err := session.NewSession(awsConf)
	if err != nil {
		return nil, err
	}

	return &S3Backend{sess}, nil
}

// Get copies an object from S3 to the host path.
func (s3b *S3Backend) Get(ctx context.Context, url string, hostPath string, class tes.FileType) error {

	path := strings.TrimPrefix(url, S3Protocol)
	split := strings.SplitN(path, "/", 2)
	bucket := split[0]
	key := split[1]

	region, err := s3manager.GetBucketRegion(ctx, s3b.sess, bucket, "us-east-1")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			return fmt.Errorf("unable to find bucket %s's region not found: %v", bucket, aerr)
		}
		return err
	}

	// Create a downloader with the session and default options
	sess := s3b.sess.Copy(&aws.Config{Region: aws.String(region)})
	client := s3.New(sess)
	manager := s3manager.NewDownloader(sess)

	switch class {
	case File:
		// Create a file to write the S3 Object contents to.
		hf, err := os.Create(hostPath)
		if err != nil {
			return fmt.Errorf("failed to create file %q, %v", hostPath, err)
		}

		_, err = manager.DownloadWithContext(ctx, hf, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return err
		}

		err = hf.Close()
		if err != nil {
			return err
		}

	case Directory:
		err = client.ListObjectsV2PagesWithContext(
			ctx,
			&s3.ListObjectsV2Input{Bucket: &bucket, Prefix: &key},
			func(page *s3.ListObjectsV2Output, more bool) bool {
				for _, obj := range page.Contents {
					if *obj.Key != key+"/" {
						// Create the directories in the path
						file := filepath.Join(hostPath, strings.TrimPrefix(*obj.Key, key+"/"))
						if err := os.MkdirAll(filepath.Dir(file), 0775); err != nil {
							panic(err)
						}

						// Setup the local file
						hf, err := os.Create(file)
						if err != nil {
							panic(err)
						}

						// Download the file using the AWS SDK
						_, err = manager.DownloadWithContext(ctx, hf, &s3.GetObjectInput{
							Bucket: aws.String(bucket),
							Key:    obj.Key,
						})
						if err != nil {
							panic(err)
						}

						err = hf.Close()
						if err != nil {
							panic(err)
						}
					}
				}
				return true
			},
		)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}

	return nil
}

// PutFile copies an object (file) from the host path to S3.
func (s3b *S3Backend) PutFile(ctx context.Context, url string, hostPath string) error {

	path := strings.TrimPrefix(url, S3Protocol)
	split := strings.SplitN(path, "/", 2)
	bucket := split[0]
	key := split[1]

	region, err := s3manager.GetBucketRegion(ctx, s3b.sess, bucket, "us-east-1")
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
			return fmt.Errorf("unable to find bucket %s's region not found", bucket)
		}
		return err
	}

	// Create a uploader with the session and default options
	sess := s3b.sess.Copy(&aws.Config{Region: aws.String(region)})
	manager := s3manager.NewUploader(sess)

	fh, err := os.Open(hostPath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", hostPath, err)
	}
	_, err = manager.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   fh,
	})
	if err != nil {
		return err
	}

	return fh.Close()
}

// Supports indicates whether this backend supports the given storage request.
// For S3, the url must start with "s3://".
func (s3b *S3Backend) Supports(url string, hostPath string, class tes.FileType) bool {
	return strings.HasPrefix(url, S3Protocol)
}
