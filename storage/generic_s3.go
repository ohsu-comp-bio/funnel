package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/minio/minio-go"
	"github.com/ohsu-comp-bio/funnel/config"
)

// GenericS3 provides access to an S3 object store.
type GenericS3 struct {
	client   *minio.Client
	endpoint string
}

// NewGenericS3 creates a new GenericS3 instance, given an endpoint URL
// and a set of authentication credentials.
func NewGenericS3(conf config.GenericS3Storage) (*GenericS3, error) {
	ssl := strings.HasPrefix(conf.Endpoint, "https")
	endpoint := endpointRE.ReplaceAllString(conf.Endpoint, "$2")
	client, err := minio.NewV2(endpoint, conf.Key, conf.Secret, ssl)
	if err != nil {
		return nil, fmt.Errorf("error creating generic s3 backend: %v", err)
	}

	return &GenericS3{client, endpoint + "/"}, nil
}

// Stat returns information about the object at the given storage URL.
func (s3 *GenericS3) Stat(ctx context.Context, url string) (*Object, error) {
	u, err := s3.parse(url)
	if err != nil {
		return nil, err
	}

	opts := minio.GetObjectOptions{}
	obj, err := s3.client.GetObjectWithContext(ctx, u.bucket, u.path, opts)
	if err != nil {
		return nil, fmt.Errorf("generic s3: getting object: %s", err)
	}

	info, err := obj.Stat()
	if err != nil {
		return nil, fmt.Errorf("generic s3: stat object: %s", err)
	}

	return &Object{
		URL:          url,
		Name:         info.Key,
		ETag:         info.ETag,
		LastModified: info.LastModified,
		Size:         info.Size,
	}, nil
}

// List lists the objects at the given url.
func (s3 *GenericS3) List(ctx context.Context, url string) ([]*Object, error) {
	u, err := s3.parse(url)
	if err != nil {
		return nil, err
	}

	// Recursively list all objects.
	var objects []*Object
	recursive := true
	for info := range s3.client.ListObjects(u.bucket, u.path, recursive, ctx.Done()) {
		// check if key represents a directory
		if strings.HasSuffix(info.Key, "/") {
			continue
		}
		objects = append(objects, &Object{
			URL:          "s3://" + u.bucket + "/" + info.Key,
			Name:         info.Key,
			ETag:         info.ETag,
			LastModified: info.LastModified,
			Size:         info.Size,
		})
	}

	return objects, nil
}

// Get copies an object from S3 to the host path.
func (s3 *GenericS3) Get(ctx context.Context, url, path string) (*Object, error) {
	obj, err := s3.Stat(ctx, url)
	if err != nil {
		return nil, err
	}

	u, err := s3.parse(url)
	if err != nil {
		return nil, err
	}

	opts := minio.GetObjectOptions{}
	err = s3.client.FGetObjectWithContext(ctx, u.bucket, u.path, path, opts)
	if err != nil {
		return nil, fmt.Errorf("generic s3: getting object: %s", err)
	}
	return obj, nil
}

// Put copies an object (file) from the host path to S3.
func (s3 *GenericS3) Put(ctx context.Context, url, path string) (*Object, error) {
	u, err := s3.parse(url)
	if err != nil {
		return nil, err
	}

	opts := minio.PutObjectOptions{}
	_, err = s3.client.FPutObjectWithContext(ctx, u.bucket, u.path, path, opts)
	if err != nil {
		return nil, err
	}
	return s3.Stat(ctx, url)
}

// Join joins the given URL with the given subpath.
func (s3 *GenericS3) Join(url, path string) (string, error) {
	if path == "" {
		return url, nil
	}
	return strings.TrimSuffix(url, "/") + "/" + path, nil
}

// UnsupportedOperations describes which operations (Get, Put, etc) are not
// supported for the given URL.
func (s3 *GenericS3) UnsupportedOperations(url string) UnsupportedOperations {
	u, err := s3.parse(url)
	if err != nil {
		return AllUnsupported(err)
	}
	ok, err := s3.client.BucketExists(u.bucket)
	if err != nil {
		err = fmt.Errorf("genericS3: failed to find bucket %q: %v", u.bucket, err)
		return AllUnsupported(err)
	}
	if !ok {
		err := fmt.Errorf("genericS3: bucket does not exist: %q", u.bucket)
		return AllUnsupported(err)
	}
	return AllSupported()
}

func (s3 *GenericS3) parse(rawurl string) (*urlparts, error) {
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
