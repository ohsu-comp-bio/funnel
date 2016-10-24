package tesTaskengineWorker

import (
	"fmt"
	"log"
	"strings"
)

// FileProtocol documentation
// TODO: documentation
var FileProtocol = "file://"

// FileAccess documentation
// TODO: documentation
type FileAccess struct {
	Allowed []string
}

// NewFileAccess documentation
// TODO: documentation
func NewFileAccess(allowed []string) *FileAccess {
	return &FileAccess{Allowed: allowed}
}

// Get documentation
// TODO: documentation
func (fileAccess *FileAccess) Get(storage string, hostPath string, class string) error {
	log.Printf("Starting download of %s", storage)
	storage = strings.TrimPrefix(storage, FileProtocol)
	found := false
	for _, i := range fileAccess.Allowed {
		if strings.HasPrefix(storage, i) {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("Can't access file %s", storage)
	}
	if class == "File" {
		CopyFile(storage, hostPath)
		return nil
	} else if class == "Directory" {
		CopyDir(storage, hostPath)
		return nil
	}
	return fmt.Errorf("Unknown element type: %s", class)

}

// Put documentation
// TODO: documentation
func (fileAccess *FileAccess) Put(storage string, hostPath string, class string) error {
	log.Printf("Starting upload of %s", storage)
	storage = strings.TrimPrefix(storage, FileProtocol)
	found := false
	for _, i := range fileAccess.Allowed {
		if strings.HasPrefix(storage, i) {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("Can't access file %s", storage)
	}
	if class == "File" {
		CopyFile(hostPath, storage)
		return nil
	} else if class == "Directory" {
		CopyDir(hostPath, storage)
		return nil
	}
	return fmt.Errorf("Unknown element type: %s", class)
}
