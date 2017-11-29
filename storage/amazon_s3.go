package storage

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	util "github.com/ohsu-comp-bio/funnel/util/aws"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var endpointRegExp = regexp.MustCompile("^(http[s]?://)?(.[^/]+)(/+)?$")

// s3Protocol defines the s3 URL protocol
const s3Protocol = "s3://"

// AmazonS3Backend provides access to an S3 object store.
type AmazonS3Backend struct {
	sess     *session.Session
	endpoint string
}

// NewAmazonS3Backend creates an AmazonS3Backend session instance
func NewAmazonS3Backend(conf config.AmazonS3Storage) (*AmazonS3Backend, error) {
	sess, err := util.NewAWSSession(conf.AWS)
	if err != nil {
		return nil, err
	}

	endpoint := endpointRegExp.ReplaceAllString(conf.AWS.Endpoint, "$2/")

	return &AmazonS3Backend{sess, endpoint}, nil
}

// Get copies an object from S3 to the host path.
func (s3b *AmazonS3Backend) Get(ctx context.Context, rawurl string, hostPath string, class tes.FileType) error {

	url := s3b.parse(rawurl)

	region, err := s3manager.GetBucketRegion(ctx, s3b.sess, url.bucket, "us-east-1")
	if err != nil {
		return fmt.Errorf("failed to determine region for bucket: %s. error: %v", url.bucket, err)
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
			Bucket: aws.String(url.bucket),
			Key:    aws.String(url.path),
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
			&s3.ListObjectsV2Input{Bucket: aws.String(url.bucket), Prefix: aws.String(url.path)},
			func(page *s3.ListObjectsV2Output, more bool) bool {
				for _, obj := range page.Contents {
					if *obj.Key != url.path+"/" {
						// Create the directories in the path
						file := filepath.Join(hostPath, strings.TrimPrefix(*obj.Key, url.path+"/"))
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
							Bucket: aws.String(url.bucket),
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
func (s3b *AmazonS3Backend) PutFile(ctx context.Context, rawurl string, hostPath string) error {

	url := s3b.parse(rawurl)

	region, err := s3manager.GetBucketRegion(ctx, s3b.sess, url.bucket, "us-east-1")
	if err != nil {
		return fmt.Errorf("failed to determine region for bucket: %s. error: %v", url.bucket, err)
	}

	// Create a uploader with the session and default options
	sess := s3b.sess.Copy(&aws.Config{Region: aws.String(region)})
	manager := s3manager.NewUploader(sess)

	fh, err := os.Open(hostPath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", hostPath, err)
	}
	_, err = manager.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(url.bucket),
		Key:    aws.String(url.path),
		Body:   fh,
	})
	if err != nil {
		return err
	}

	return fh.Close()
}

// Supports indicates whether this backend supports the given storage request.
// For the AmazonS3Backend, the url must start with "s3://"
func (s3b *AmazonS3Backend) Supports(rawurl string) error {
	if !strings.HasPrefix(rawurl, s3Protocol) {
		return fmt.Errorf("s3: unsupported protocol; expected %s", s3Protocol)
	}

	url := s3b.parse(rawurl)
	_, err := s3manager.GetBucketRegion(context.Background(), s3b.sess, url.bucket, "us-east-1")
	if err != nil {
		return fmt.Errorf("s3: failed to find bucket: %s. error: %v", url.bucket, err)
	}

	return nil
}

func (s3b *AmazonS3Backend) parse(url string) *urlparts {
	path := strings.TrimPrefix(url, s3Protocol)
	if s3b.endpoint != "" {
		path = strings.TrimPrefix(path, s3b.endpoint)
	} else {
		re := regexp.MustCompile("^s3.*\\.amazonaws\\.com/")
		path = re.ReplaceAllString(path, "")
	}

	split := strings.SplitN(path, "/", 2)
	bucket := split[0]
	key := split[1]

	return &urlparts{bucket, key}
}
