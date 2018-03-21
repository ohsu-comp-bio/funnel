package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/tes"
	util "github.com/ohsu-comp-bio/funnel/util/aws"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

var endpointRE = regexp.MustCompile("^(http[s]?://)?(.[^/]+)(.+)?$")

// s3Protocol defines the s3 URL protocol
const s3Protocol = "s3://"

// AmazonS3Backend provides access to an S3 object store.
type AmazonS3Backend struct {
	sess     *session.Session
	endpoint string
}

// NewAmazonS3Backend creates an AmazonS3Backend session instance
func NewAmazonS3Backend(conf config.AmazonS3Storage) (*AmazonS3Backend, error) {
	sess, err := util.NewAWSSession(conf.AWSConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating amazon s3 backend: %v", err)
	}

	var endpoint string
	if conf.Endpoint != "" {
		endpoint = endpointRE.ReplaceAllString(conf.Endpoint, "$2/")
	}

	return &AmazonS3Backend{sess, endpoint}, nil
}

// Get copies an object from S3 to the host path.
func (s3b *AmazonS3Backend) Get(ctx context.Context, rawurl string, hostPath string, class tes.FileType) (err error) {
	url, err := s3b.parse(rawurl)
	if err != nil {
		return err
	}

	region, err := s3manager.GetBucketRegion(ctx, s3b.sess, url.bucket, "us-east-1")
	if err != nil {
		return fmt.Errorf("failed to determine region for bucket: %s. error: %v", url.bucket, err)
	}

	sess := s3b.sess.Copy(&aws.Config{Region: aws.String(region)})
	client := s3.New(sess)
	manager := s3manager.NewDownloader(sess)

	switch class {
	case File:
		err = fsutil.EnsurePath(hostPath)
		if err != nil {
			return err
		}

		// Create a file to write the S3 Object contents to.
		hf, err := os.Create(hostPath)
		if err != nil {
			return err
		}
		defer func() {
			cerr := hf.Close()
			if cerr != nil {
				err = fmt.Errorf("%v; %v", err, cerr)
			}
		}()

		_, err = manager.DownloadWithContext(ctx, hf, &s3.GetObjectInput{
			Bucket: aws.String(url.bucket),
			Key:    aws.String(url.path),
		})
		if err != nil {
			return err
		}

	case Directory:
		err = fsutil.EnsureDir(hostPath)
		if err != nil {
			return err
		}

		objects := []*s3.Object{}

		err = client.ListObjectsV2PagesWithContext(
			ctx,
			&s3.ListObjectsV2Input{Bucket: aws.String(url.bucket), Prefix: aws.String(url.path)},
			func(page *s3.ListObjectsV2Output, more bool) bool {
				objects = append(objects, page.Contents...)
				return true
			},
		)
		if err != nil {
			return err
		}
		if len(objects) == 0 {
			return ErrEmptyDirectory
		}

		for _, obj := range objects {
			if *obj.Key != url.path+"/" {
				file := filepath.Join(hostPath, strings.TrimPrefix(*obj.Key, url.path+"/"))
				// check if key represents a directory
				if strings.HasSuffix(*obj.Key, "/") {
					continue
				}
				err = fsutil.EnsurePath(file)
				if err != nil {
					return err
				}

				// Setup the local file
				hf, err := os.Create(file)
				if err != nil {
					return err
				}

				// Download the file using the AWS SDK
				_, err = manager.DownloadWithContext(ctx, hf, &s3.GetObjectInput{
					Bucket: aws.String(url.bucket),
					Key:    obj.Key,
				})
				if err != nil {
					cerr := hf.Close()
					if cerr != nil {
						return fmt.Errorf("%v; %v", err, cerr)
					}
					return err
				}

				err = hf.Close()
				if err != nil {
					return err
				}
			}
		}

	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}

	return nil
}

// PutFile copies an object (file) from the host path to S3.
func (s3b *AmazonS3Backend) PutFile(ctx context.Context, rawurl string, hostPath string) error {
	url, err := s3b.parse(rawurl)
	if err != nil {
		return err
	}

	region, err := s3manager.GetBucketRegion(ctx, s3b.sess, url.bucket, "us-east-1")
	if err != nil {
		return fmt.Errorf("failed to determine region for bucket: %s. error: %v", url.bucket, err)
	}

	sess := s3b.sess.Copy(&aws.Config{Region: aws.String(region)})
	manager := s3manager.NewUploader(sess)

	hf, err := os.Open(hostPath)
	if err != nil {
		return fmt.Errorf("failed to open file %q, %v", hostPath, err)
	}
	defer hf.Close()

	_, err = manager.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(url.bucket),
		Key:    aws.String(url.path),
		Body:   hf,
	})

	return err
}

// SupportsGet indicates whether this backend supports GET storage request.
// For the AmazonS3Backend, the url must start with "s3://" and the bucket must exist
func (s3b *AmazonS3Backend) SupportsGet(rawurl string, class tes.FileType) error {
	url, err := s3b.parse(rawurl)
	if err != nil {
		return err
	}

	_, err = s3manager.GetBucketRegion(context.Background(), s3b.sess, url.bucket, "us-east-1")
	if err != nil {
		return fmt.Errorf("amazonS3: failed to determine region for bucket: %s. error: %v", url.bucket, err)
	}

	return nil
}

// SupportsPut indicates whether this backend supports PUT storage request.
// For the AmazonS3Backend, the url must start with "s3://" and the bucket must exist
func (s3b *AmazonS3Backend) SupportsPut(rawurl string, class tes.FileType) error {
	return s3b.SupportsGet(rawurl, class)
}

func (s3b *AmazonS3Backend) parse(rawurl string) (*urlparts, error) {
	if !strings.HasPrefix(rawurl, s3Protocol) {
		return nil, &ErrUnsupportedProtocol{"amazonS3"}
	}

	path := strings.TrimPrefix(rawurl, s3Protocol)
	if s3b.endpoint != "" {
		path = strings.TrimPrefix(path, s3b.endpoint)
	} else {
		re := regexp.MustCompile("^s3.*\\.amazonaws\\.com/")
		path = re.ReplaceAllString(path, "")
	}
	if path == "" {
		return nil, &ErrInvalidURL{"amazonS3"}
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
