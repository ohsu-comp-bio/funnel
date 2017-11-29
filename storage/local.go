package storage

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
)

// LocalBackend provides access to a local-disk storage system.
type LocalBackend struct {
	allowedDirs []string
}

// NewLocalBackend returns a LocalBackend instance, configured to limit
// file system access to the given allowed directories.
func NewLocalBackend(conf config.LocalStorage) (*LocalBackend, error) {
	allowed := []string{}
	for _, d := range conf.AllowedDirs {
		a, err := filepath.Abs(d)
		if err != nil {
			return nil, err
		}
		allowed = append(allowed, a)
	}
	return &LocalBackend{allowed}, nil
}

// Get copies a file from storage into the given hostPath.
func (local *LocalBackend) Get(ctx context.Context, url string, hostPath string, class tes.FileType) error {
	path := getPath(url)

	var err error
	if class == File {
		err = linkFile(path, hostPath)
	} else if class == Directory {
		err = filepath.Walk(path, func(p string, f os.FileInfo, err error) error {
			if !f.IsDir() {
				rel, err := filepath.Rel(path, p)
				if err != nil {
					return err
				}
				return local.Get(ctx, p, filepath.Join(hostPath, rel), File)
			}
			return nil
		})
	} else {
		err = fmt.Errorf("Unknown file class: %s", class)
	}

	if err != nil {
		return err
	}
	return nil
}

// PutFile copies a file from the hostPath into storage.
func (local *LocalBackend) PutFile(ctx context.Context, url string, hostPath string) error {
	path := getPath(url)
	return linkFile(hostPath, path)
}

// Supports indicates whether this backend supports the given storage request.
// For the LocalBackend, the url must start with "file://" be in an allowed directory
func (local *LocalBackend) Supports(rawurl string) error {
	path := getPath(rawurl)
	ok := strings.HasPrefix(path, "/")
	if !ok {
		return fmt.Errorf("local: must provide an absolute path: %s", rawurl)
	}
	if !isAllowed(path, local.allowedDirs) {
		return fmt.Errorf("local: can't access file, path is not in allowed directories: %s", rawurl)
	}
	return nil
}

func getPath(rawurl string) string {
	p := strings.TrimPrefix(rawurl, "file://")
	return p
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
func copyFile(source string, dest string) (err error) {
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
		return err
	}
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	_, err = io.Copy(df, sf)
	if err != nil {
		return err
	}
	// close files
	err = sf.Close()
	if err != nil {
		return err
	}
	err = df.Close()
	if err != nil {
		return err
	}
	return nil
}

// Hard links file source to destination dest.
func linkFile(source string, dest string) error {
	var err error
	// without this resulting link could be a symlink
	parent, err := filepath.EvalSymlinks(source)
	if err != nil {
		return err
	}
	same, err := sameFile(parent, dest)
	if err != nil {
		return err
	}
	if same {
		return nil
	}
	// make parent dirs if they dont exist
	dstD := path.Dir(dest)
	if _, err := os.Stat(dstD); err != nil {
		_ = syscall.Umask(0000)
		err = os.MkdirAll(dstD, 0775)
		if err != nil {
			return err
		}
	}
	err = os.Link(parent, dest)
	if err != nil {
		err = copyFile(source, dest)
	}
	return err
}

func sameFile(source string, dest string) (bool, error) {
	var same bool
	var err error
	sfi, err := os.Stat(source)
	if err != nil {
		return same, err
	}
	dfi, err := os.Stat(dest)
	if os.IsNotExist(err) {
		return same, nil
	} else if err != nil {
		return same, err
	}
	same = os.SameFile(sfi, dfi)
	return same, nil
}
