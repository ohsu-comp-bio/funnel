package storage

import (
	"context"
	"fmt"
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io"
	"io/ioutil"
	"net/url"
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
	return &LocalBackend{conf.AllowedDirs}, nil
}

// Get copies a file from storage into the given hostPath.
func (local *LocalBackend) Get(ctx context.Context, url string, hostPath string, class tes.FileType) error {
	log.Info("Starting download", "url", url, "hostPath", hostPath)
	path, ok := getPath(url)

	if !ok {
		return fmt.Errorf("local storage does not support put on %s", url)
	}

	if !isAllowed(path, local.allowedDirs) {
		return fmt.Errorf("Can't access file, path is not in allowed directories:  %s", path)
	}

	var err error
	if class == File {
		err = linkFile(path, hostPath)
	} else if class == Directory {
		err = copyDir(path, hostPath)
	} else {
		err = fmt.Errorf("Unknown file class: %s", class)
	}

	if err == nil {
		log.Info("Finished download", "url", url, "hostPath", hostPath)
	}
	return err
}

// Put copies a file from the hostPath into storage.
func (local *LocalBackend) Put(ctx context.Context, url string, hostPath string, class tes.FileType) error {
	log.Info("Starting upload", "url", url, "hostPath", hostPath)
	path, ok := getPath(url)

	if !ok {
		return fmt.Errorf("local storage does not support put on %s", url)
	}

	if !isAllowed(path, local.allowedDirs) {
		return fmt.Errorf("Can't access file, path is not in allowed directories:  %s", url)
	}

	if class == File {
		err := linkFile(hostPath, path)
		if err != nil {
			return err
		}
	} else if class == Directory {
		err := copyDir(hostPath, path)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Unknown file class: %s", class)
	}
	log.Info("Finished upload", "url", url, "hostPath", hostPath)
	return nil
}

// Supports indicates whether this backend supports the given storage request.
// For the LocalBackend, the url must start with "file://"
func (local *LocalBackend) Supports(rawurl string, hostPath string, class tes.FileType) bool {
	_, ok := getPath(rawurl)
	return ok
}

func getPath(rawurl string) (string, bool) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return "", false
	}
	if u.Path == "" {
		return "", false
	}
	if u.Scheme == "file" {
		return u.Path, true
	}
	// Handle URLs that are file paths, e.g. "/path/to/foo.txt"
	if u.Scheme == "" && u.Host == "" {
		return u.Path, true
	}
	return "", false
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
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	dstD := path.Dir(dest)
	// make parent dirs if they dont exist
	if _, err := os.Stat(dstD); err != nil {
		_ = syscall.Umask(0000)
		os.MkdirAll(dstD, 0777)
	}
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	_, err = io.Copy(df, sf)
	cerr := df.Close()
	if err != nil {
		return err
	}
	if cerr != nil {
		return cerr
	}
	// ensure readable output files
	err = os.Chmod(dest, 0666)
	if err != nil {
		return err
	}
	return nil
}

// Recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must *not* exist.
func copyDir(source string, dest string) (err error) {
	// get properties of source dir
	fi, err := os.Stat(source)
	if err != nil {
		return err
	}

	if !fi.IsDir() {
		return fmt.Errorf("Source is not a directory")
	}

	// ensure dest dir does not already exist
	_, err = os.Open(dest)
	if os.IsNotExist(err) {
		// create dest dir
		_ = syscall.Umask(0000)
		err = os.MkdirAll(dest, 0777)
		if err != nil {
			return err
		}
	} else if err != nil {
		log.Error("copyDir os.Open error", err)
		return err
	}

	entries, err := ioutil.ReadDir(source)
	for _, entry := range entries {
		sfp := source + "/" + entry.Name()
		dfp := dest + "/" + entry.Name()
		if entry.IsDir() {
			err = copyDir(sfp, dfp)
			if err != nil {
				return err
			}
		} else {
			// create hard link; falls back to copy on error
			err = linkFile(sfp, dfp)
			if err != nil {
				return err
			}
		}
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
	err = os.Link(parent, dest)
	if err != nil {
		log.Debug("Failed to link file; attempting copy",
			"linkErr", err,
			"source", source,
			"dest", dest)
		err = copyFile(source, dest)
	}
	return err
}
