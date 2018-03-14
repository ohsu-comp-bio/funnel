package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util/fsutil"
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

	switch class {
	case File:
		err = fsutil.EnsurePath(hostPath)
		if err != nil {
			return err
		}
		err = linkFile(ctx, path, hostPath)

	case Directory:
		files, err := walkFiles(path)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			return ErrEmptyDirectory
		}

		for _, f := range files {
			p := filepath.Join(hostPath, f.rel)
			err := local.Get(ctx, f.abs, p, File)
			if err != nil {
				return err
			}
		}

	default:
		err = fmt.Errorf("Unknown file class: %s", class)
	}

	return err
}

// PutFile copies a file from the hostPath into storage.
func (local *LocalBackend) PutFile(ctx context.Context, url string, hostPath string) error {
	path := getPath(url)
	err := fsutil.EnsurePath(path)
	if err != nil {
		return err
	}
	return linkFile(ctx, hostPath, path)
}

// SupportsGet indicates whether this backend supports GET storage request.
// For the LocalBackend, the url must start with "file://" be in an allowed directory
func (local *LocalBackend) SupportsGet(rawurl string, class tes.FileType) error {
	if !strings.HasPrefix(rawurl, "/") && !strings.HasPrefix(rawurl, "file://") {
		return &ErrUnsupportedProtocol{"localStorage"}
	}

	path := getPath(rawurl)
	if !isAllowed(path, local.allowedDirs) {
		return fmt.Errorf("localStorage: can't access file, path is not in allowed directories: %s", rawurl)
	}
	return nil
}

// SupportsPut indicates whether this backend supports PUT storage request.
// For the LocalBackend, the url must start with "file://" be in an allowed directory
func (local *LocalBackend) SupportsPut(rawurl string, class tes.FileType) error {
	return local.SupportsGet(rawurl, class)
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
		return fmt.Errorf("failed to open source file for copying: %s", err)
	}
	defer sf.Close()

	// Create and open dest file for writing
	df, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 0775)
	if err != nil {
		return fmt.Errorf("failed to create dest file for copying: %s", err)
	}
	defer func() {
		cerr := df.Close()
		if cerr != nil {
			err = fmt.Errorf("%v; %v", err, cerr)
		}
	}()

	_, err = fsutil.Copy(ctx, df, sf)
	return err
}

// Hard links file source to destination dest.
func linkFile(ctx context.Context, source string, dest string) error {
	var err error
	// without this resulting link could be a symlink
	parent, err := filepath.EvalSymlinks(source)
	if err != nil {
		return fmt.Errorf("failed to eval symlinks: %s", err)
	}
	same, err := sameFile(parent, dest)
	if err != nil {
		return fmt.Errorf("failed to check if file is the same file: %s", err)
	}
	if same {
		return nil
	}
	err = os.Link(parent, dest)
	if err != nil {
		err = copyFile(ctx, source, dest)
		if err != nil {
			return fmt.Errorf("failed to copy file: %s", err)
		}
	}
	return err
}

func sameFile(source string, dest string) (bool, error) {
	var err error
	sfi, err := os.Stat(source)
	if err != nil {
		return false, fmt.Errorf("failed to stat src file: %s", err)
	}
	dfi, err := os.Stat(dest)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to stat dest file: %s", err)
	}
	return os.SameFile(sfi, dfi), nil
}
