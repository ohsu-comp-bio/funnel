package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	pathlib "path"
	"path/filepath"
	"strings"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

// Local provides access to a local-disk storage system.
type Local struct {
	allowedDirs []string
}

// NewLocal returns a Local instance, configured to limit
// file system access to the given allowed directories.
func NewLocal(conf config.LocalStorage) (*Local, error) {
	allowed := []string{}
	for _, d := range conf.AllowedDirs {
		a, err := filepath.Abs(d)
		if err != nil {
			return nil, err
		}
		allowed = append(allowed, a)
	}
	return &Local{allowed}, nil
}

// Stat returns information about the object at the given storage URL.
func (local *Local) Stat(ctx context.Context, url string) (*Object, error) {
	path := getPath(url)
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return &Object{
		URL:          url,
		Name:         path,
		LastModified: info.ModTime(),
		Size:         info.Size(),
	}, nil
}

// List lists the objects at the given url.
func (local *Local) List(ctx context.Context, url string) ([]*Object, error) {
	files, err := fsutil.WalkFiles(getPath(url))
	if err != nil {
		return nil, err
	}

	var objects []*Object
	for _, f := range files {
		url, err := local.Join(url, f.Rel)
		if err != nil {
			return nil, err
		}
		objects = append(objects, &Object{
			URL:          url,
			Name:         f.Rel,
			LastModified: f.LastModified,
			Size:         f.Size,
		})
	}
	return objects, nil
}

// Get copies a file from storage into the given hostPath.
func (local *Local) Get(ctx context.Context, url, path string) (*Object, error) {
	err := linkFile(ctx, getPath(url), path)
	if err != nil {
		return nil, err
	}
	return local.Stat(ctx, url)
}

// Put copies a file from the hostPath into storage.
func (local *Local) Put(ctx context.Context, url, path string) (*Object, error) {
	target := getPath(url)
	err := fsutil.EnsurePath(target)
	if err != nil {
		return nil, err
	}

	err = linkFile(ctx, path, target)
	if err != nil {
		return nil, err
	}
	return local.Stat(ctx, url)
}

// Join joins the given URL with the given subpath.
func (local *Local) Join(url, path string) (string, error) {
	if strings.HasPrefix(url, "file://") {
		return "file://" + pathlib.Join(strings.TrimPrefix(url, "file://"), path), nil
	}
	return filepath.Join(url, path), nil
}

// UnsupportedOperations describes which operations (Get, Put, etc) are not
// supported for the given URL.
func (local *Local) UnsupportedOperations(url string) UnsupportedOperations {
	if !strings.HasPrefix(url, "/") && !strings.HasPrefix(url, "file://") {
		return AllUnsupported(&ErrUnsupportedProtocol{"localStorage"})
	}

	path := getPath(url)
	if !isAllowed(path, local.allowedDirs) {
		err := fmt.Errorf(
			"localStorage: can't access file, path is not in allowed directories: %s", url)
		return AllUnsupported(err)
	}
	return AllSupported()
}

func getPath(rawurl string) string {
	return strings.TrimPrefix(rawurl, "file://")
}

func isAllowed(path string, allowedDirs []string) bool {
	for _, dir := range allowedDirs {
		if strings.HasPrefix(path, dir) {
			return true
		}
	}
	return false
}

// Copies file source to destination dest.
func copyFile(ctx context.Context, source string, dest string) (err error) {
	// check if dest exists; if it does check if it is the same as the source
	same, err := sameFile(source, dest)
	if err != nil {
		return err
	}
	if same {
		return nil
	}
	// Open source file for copying
	sf, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("failed to open source file for copying: %v", err)
	}
	defer sf.Close()

	// Create and open dest file for writing
	df, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		return fmt.Errorf("failed to create dest file for copying: %v", err)
	}

	_, copyErr := io.Copy(df, fsutil.Reader(ctx, sf))
	closeErr := df.Close()
	if copyErr != nil {
		return fmt.Errorf("copying file: %v", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("closing file: %v", closeErr)
	}

	return err
}

// Hard links file source to destination dest.
func linkFile(ctx context.Context, source string, dest string) error {
	var err error
	// without this resulting link could be a symlink
	parent, err := filepath.EvalSymlinks(source)
	if err != nil {
		return fmt.Errorf("failed to eval symlinks: %v", err)
	}
	same, err := sameFile(parent, dest)
	if err != nil {
		return fmt.Errorf("failed to check if file is the same file: %v", err)
	}
	if same {
		return nil
	}
	err = os.Link(parent, dest)
	if err != nil {
		err = copyFile(ctx, source, dest)
		if err != nil {
			return fmt.Errorf("failed to copy file: %v", err)
		}
	}
	return err
}

func sameFile(source string, dest string) (bool, error) {
	var err error
	sfi, err := os.Stat(source)
	if err != nil {
		return false, fmt.Errorf("failed to stat src file: %v", err)
	}
	dfi, err := os.Stat(dest)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to stat dest file: %v", err)
	}
	return os.SameFile(sfi, dfi), nil
}
