package ga4gh_taskengine_worker

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
)

type FileStorageAccess struct {
	StorageDir string
}

func NewSharedFS(base string) *FileStorageAccess {
	return &FileStorageAccess{StorageDir: base}
}

func (self *FileStorageAccess) Get(storage string, hostPath string) error {
	storage = strings.TrimPrefix(storage, "fs://")
	srcPath := path.Join(self.StorageDir, storage)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("storage file '%s' not found", srcPath)
	}
	copyFileContents(srcPath, hostPath)
	return nil
}

func (self *FileStorageAccess) Put(location string, hostPath string, class string) error {

	storage := strings.TrimPrefix(location, "fs://")

	log.Printf("copy out %s %s\n", hostPath, path.Join(self.StorageDir, storage))
	//copy to storage directory
	if class == "Directory" {
		err := CopyDir(hostPath, path.Join(self.StorageDir, storage))
		if err != nil {
			log.Printf("Error copying output directory %s to %s", hostPath, location)
			return err
		}
	} else if class == "File" {
		err := CopyFile(hostPath, path.Join(self.StorageDir, storage))
		if err != nil {
			log.Printf("Error copying output file %s to %s", hostPath, location)
			return err
		}
	} else {
		return fmt.Errorf("Unknown Class type: %s", class)
	}
	return nil
}
