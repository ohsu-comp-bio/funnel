package tes_taskengine_worker

import (
	"fmt"
	"log"
	"strings"
)

var FILE_PROTOCOL = "file://"

type FileAccess struct {
	Allowed []string
}

func NewFileAccess(allowed []string) *FileAccess {
	return &FileAccess{Allowed: allowed}
}

func (self *FileAccess) Get(storage string, hostPath string, class string) error {
	log.Printf("Starting download of %s", storage)
	storage = strings.TrimPrefix(storage, FILE_PROTOCOL)
	found := false
	for _, i := range self.Allowed {
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

func (self *FileAccess) Put(storage string, hostPath string, class string) error {
	log.Printf("Starting upload of %s", storage)
	storage = strings.TrimPrefix(storage, FILE_PROTOCOL)
	found := false
	for _, i := range self.Allowed {
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
