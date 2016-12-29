package storage

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
)

// Protocol defines the expected prefix of URI matching this storage system.
// e.g. "file:///path/to/file" matches the Local storage system.
const LocalProtocol = "file://"

// LocalBackend provides access to a local-disk storage system.
type LocalBackend struct {
	allowedDirs []string
}

// NewLocalBackend returns a LocalBackend instance, configured to limit
// file system access to the given allowed directories.
func NewLocalBackend(allowed []string) *LocalBackend {
	return &LocalBackend{allowed}
}

// Get copies a file from storage into the given hostPath.
func (local *LocalBackend) Get(url string, hostPath string, class string) error {
	log.Printf("Starting download of local file: %s", url)
	path := strings.TrimPrefix(url, LocalProtocol)

	if !isAllowed(path, local.allowedDirs) {
		return fmt.Errorf("Can't access file, path is not in allowed directories:  %s", path)
	}

	if class == File {
		copyFile(path, hostPath)
	} else if class == Directory {
		copyDir(path, hostPath)
	} else {
		return fmt.Errorf("Unknown file class: %s", class)
	}
	log.Printf("Finished download of local file: %s", url)
	return nil
}

// Put copies a file from the hostPath into storage.
func (local *LocalBackend) Put(url string, hostPath string, class string) error {
	log.Printf("Starting upload to local file: %s", url)
	path := strings.TrimPrefix(url, LocalProtocol)

	if !isAllowed(path, local.allowedDirs) {
		return fmt.Errorf("Can't access file, path is not in allowed directories:  %s", url)
	}

	if class == File {
		copyFile(hostPath, path)
	} else if class == Directory {
		copyDir(hostPath, path)
	} else {
		return fmt.Errorf("Unknown file class: %s", class)
	}
	log.Printf("Finished upload to local file: %s", url)
	return nil
}

// Determines whether this backend matches the given url
func (s *LocalBackend) Supports(url string, hostPath string, class string) bool {
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

func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	err = out.Sync()
	return nil
}

func copyFile(src, dst string) (err error) {
	sfi, err := os.Stat(src)
	if err != nil {
		return
	}
	if !sfi.Mode().IsRegular() {
		// This cannot copy non-regular files (e.g.,
		// directories, symlinks, devices, etc.)
		return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
		dstD := path.Dir(dst)
		if _, err := os.Stat(dstD); err != nil {
			fmt.Printf("Making %s\n", dstD)
			os.MkdirAll(dstD, 0700)
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}

		if os.SameFile(sfi, dfi) {
			return
		}
	}

	err = copyFileContents(src, dst)
	return
}

func copyDir(source string, dest string) (err error) {
	// Gets properties of source directory.
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// Creates destination directory.
	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)
	objects, err := directory.Readdir(-1)
	for _, obj := range objects {
		sourcefilepointer := source + "/" + obj.Name()
		destinationfilepointer := dest + "/" + obj.Name()
		if obj.IsDir() {
			// Creates sub-directories recursively.
			err = copyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// Performs copy.
			err = copyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}
