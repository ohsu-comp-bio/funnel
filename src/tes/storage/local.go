package storage

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"
	"tes/config"
)

// LocalProtocol defines the expected prefix of URL matching this storage system.
// e.g. "file:///path/to/file" matches the Local storage system.
const LocalProtocol = "file://"

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
func (local *LocalBackend) Get(ctx context.Context, url string, hostPath string, class string) error {
	log.Info("Starting download", "url", url, "hostPath", hostPath, "class", class)
	path := strings.TrimPrefix(url, LocalProtocol)

	if !isAllowed(path, local.allowedDirs) {
		return fmt.Errorf("Can't access file, path is not in allowed directories:  %s", path)
	}
	switch class {
	case File:
		err := copyFile(path, hostPath)
		if err != nil {
			return err
		}
	case ReadOnlyFile:
		err := linkFile(path, hostPath)
		if err != nil {
			return err
		}
	case Directory:
		err := copyDir(path, hostPath)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}

	log.Info("Finished download", "url", url, "hostPath", hostPath)
	return nil
}

// Put copies a file from the hostPath into storage.
func (local *LocalBackend) Put(ctx context.Context, url string, hostPath string, class string) error {
	log.Info("Starting upload", "url", url, "hostPath", hostPath)
	path := strings.TrimPrefix(url, LocalProtocol)

	if !isAllowed(path, local.allowedDirs) {
		return fmt.Errorf("Can't access file, path is not in allowed directories:  %s", url)
	}

	switch class {
	// Outputs are always copied
	case File, ReadOnlyFile:
		err := copyFile(hostPath, path)
		if err != nil {
			return err
		}
	case Directory:
		err := copyDir(hostPath, path)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unknown file class: %s", class)
	}

	log.Info("Finished upload", "url", url, "hostPath", hostPath)
	return nil
}

// Supports indicates whether this backend supports the given storage request.
// For the LocalBackend, the url must start with "file://"
func (local *LocalBackend) Supports(url string, hostPath string, class string) bool {
	return strings.HasPrefix(url, LocalProtocol)
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
	same := checkSame(source, dest)
	if same {
		return nil
	}
	err = precheck(source, dest)
	if err != nil {
		return err
	}
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
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
	// ensure readable files
	err = os.Chmod(dest, 0666)
	if err != nil {
		return err
	}
	return nil
}

// Hard links file source to destination dest.
func linkFile(source string, dest string) error {
	same := checkSame(source, dest)
	if same {
		return nil
	}
	err := precheck(source, dest)
	if err != nil {
		return err
	}
	err = os.Link(source, dest)
	if err != nil {
		log.Debug("Failed to link file; attempting copy", "linkErr", err, "source", source, "dest", dest)
		err := copyFile(source, dest)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func precheck(source string, dest string) error {
	err := checkSrc(source)
	if err != nil {
		return err
	}
	err = checkDest(dest)
	if err != nil {
		return err
	}
	dstD := path.Dir(dest)
	// make parent dirs if they dont exist
	_, err = os.Stat(dstD)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		_ = syscall.Umask(0000)
		os.MkdirAll(dstD, 0777)
	}
	return nil
}

func checkSrc(source string) error {
	sfi, err := os.Lstat(source)
	if err != nil {
		return err
	}
	switch mode := sfi.Mode(); {
	case mode.IsRegular():
		return nil
	case mode&os.ModeSymlink != 0:
		_, err := os.Stat(source)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("Symlink (%s) could not be resolved", source)
			}
			return err
		}
	case !mode.IsRegular():
		// cannot copy non-regular files (e.g., directories, devices, etc.)
		return fmt.Errorf("non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	return nil
}

func checkDest(dest string) error {
	dfi, err := os.Stat(dest)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
	}
	return nil
}

func checkSame(source string, dest string) bool {
	sfi, _ := os.Stat(source)
	dfi, _ := os.Stat(dest)
	if os.SameFile(sfi, dfi) {
		return true
	}
	return false
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
			// perform copy
			err = copyFile(sfp, dfp)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
