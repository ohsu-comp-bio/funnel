package tesTaskEngineWorker

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

// FileStorageAccess documentation
// TODO: documentation
type FileStorageAccess struct {
	StorageDir string
}

// NewSharedFS documentaiton
// TODO: documentation
func NewSharedFS(base string) *FileStorageAccess {
	return &FileStorageAccess{StorageDir: base}
}

// Get copies storage into hostPath.
func (fileStorageAccess *FileStorageAccess) Get(storage string, hostPath string, class string) error {
	storage = strings.TrimPrefix(storage, "fs://")
	srcPath := path.Join(fileStorageAccess.StorageDir, storage)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("storage file '%s' not found", srcPath)
	}
	if class == "File" {
		copyFileContents(srcPath, hostPath)
	} else if class == "Directory" {
		CopyDir(srcPath, hostPath)
	} else {
		return fmt.Errorf("Unknown element type: %s", class)
	}
	return nil
}

// Put documentation
// TODO: documentation
func (fileStorageAccess *FileStorageAccess) Put(location string, hostPath string, class string) error {

	storage := strings.TrimPrefix(location, "fs://")

	log.Printf("copy out %s %s\n", hostPath, path.Join(fileStorageAccess.StorageDir, storage))
	// Copies to storage directory.
	if class == "Directory" {
		err := CopyDir(hostPath, path.Join(fileStorageAccess.StorageDir, storage))
		if err != nil {
			log.Printf("Error copying output directory %s to %s", hostPath, location)
			return err
		}
	} else if class == "File" {
		err := CopyFile(hostPath, path.Join(fileStorageAccess.StorageDir, storage))
		if err != nil {
			log.Printf("Error copying output file %s to %s", hostPath, location)
			return err
		}
	} else {
		return fmt.Errorf("Unknown Class type: %s", class)
	}
	return nil
}
