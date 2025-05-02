package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go"
	"github.com/ohsu-comp-bio/funnel/config"
	s3util "github.com/ohsu-comp-bio/funnel/util/s3"
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

	var endpoint string
	if conf.Endpoint != "" {
		endpoint = s3util.ParseEndpoint(conf.Endpoint)
	}

	fmt.Println("DEBUG: endpoint:", endpoint)

	client, err := minio.NewV2(endpoint, conf.Key, conf.Secret, ssl)
	if err != nil {
		return nil, fmt.Errorf("error creating generic s3 backend: %v", err)
	}

	return &GenericS3{client, endpoint + "/"}, nil
}

// Returns true if a remote S3 object is a directory, false otherwise
func isDir(minioClient *minio.Client, bucketName, objectName string) (bool, error) {
	// Check if the objectName ends with '/' - often used to represent 'folders'
	if strings.HasSuffix(objectName, "/") {
		return true, nil
	}

	// List objects with the prefix to see if there are multiple keys with the given prefix
	// Create a done channel.
	doneCh := make(chan struct{})
	defer close(doneCh)

	// Recursively list all objects
	recursive := true
	for object := range minioClient.ListObjects(bucketName, objectName, recursive, doneCh) {
		if object.Err != nil {
			return false, object.Err
		}

		// If any object's key starts with the objectName and is not equal, it's a directory
		if strings.HasPrefix(object.Key, objectName) && object.Key != objectName {
			return true, nil
		}
	}

	// If no objects share the prefix or the objectName exactly matches a key, it's considered a file
	return false, nil
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
		return nil, fmt.Errorf("genericS3: getting object: %s", err)
	}

	isDir, err := isDir(s3.client, u.bucket, u.path)
	if err != nil {
		return nil, fmt.Errorf("genericS3: stat object: %s", err)
	}
	if isDir {
		return &Object{
			URL:  url,
			Name: u.path,
			Size: 0,
		}, nil
	}

	info, err := obj.Stat()
	if err != nil {
		return nil, fmt.Errorf("genericS3: stat object: %s", err)
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

	isDir, err := isDir(s3.client, u.bucket, u.path)
	if err != nil {
		return nil, fmt.Errorf("genericS3: getting object %s: %v", url, err)
	}
	if isDir {
		objects, err := s3.List(ctx, url)
		if err != nil {
			return nil, err
		}

		for _, obj := range objects {
			// Recursively download files and subdirectories
			_, err := s3.Get(ctx, obj.URL, filepath.Join(path, obj.Name))
			if err != nil {
				return nil, err
			}
		}
	} else {
		opts := minio.GetObjectOptions{}
		// err = s3.client.FGetObjectWithContext(ctx, u.bucket, u.path, path, opts)
		err = download(ctx, s3.client, u.bucket, u.path, path, opts)
		if err != nil {
			return nil, fmt.Errorf("genericS3: getting object %s: %v", url, err)
		}
	}

	return obj, nil
}

// download streams an object to a file without using os.Rename
func download(ctx context.Context, client *minio.Client, bucket, objectPath, filePath string, opts minio.GetObjectOptions) error {
	// Step 1: Get the object stream
	obj, err := client.GetObjectWithContext(ctx, bucket, objectPath, opts)
	if err != nil {
		return fmt.Errorf("getting object from S3: %w", err)
	}
	defer obj.Close()

	// Step 2: Create the local file (overwrite if exists)
	outFile, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer outFile.Close()

	// Step 3: Copy the contents
	if _, err := io.Copy(outFile, obj); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	return nil
}

// Put copies an object (file) from the host path to S3.
// Update Put function to be able to upload directories (a la Get() function)
func (s3 *GenericS3) Put(ctx context.Context, url, path string) (*Object, error) {
	u, err := s3.parse(url)
	if err != nil {
		return nil, err
	}

	opts := minio.PutObjectOptions{}

	// Check if the path is a directory
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fileInfo.IsDir() {
		// Walk the directory and upload all files and subdirectories
		err = filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				// Upload the file
				relativePath, err := filepath.Rel(path, filePath)
				if err != nil {
					return err
				}
				uploadPath := filepath.Join(u.path, relativePath)
				_, err = s3.client.FPutObjectWithContext(ctx, u.bucket, uploadPath, filePath, opts)
				if err != nil {
					return fmt.Errorf("genericS3: putting object %s: %v", url, err)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		// Upload the file directly
		_, err = s3.client.FPutObjectWithContext(ctx, u.bucket, u.path, path, opts)
		if err != nil {
			return nil, fmt.Errorf("genericS3: putting object %s: %v", url, err)
		}
	}

	return s3.Stat(ctx, url)
}

// Join joins the given URL with the given subpath.
func (s3 *GenericS3) Join(url, path string) (string, error) {
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
