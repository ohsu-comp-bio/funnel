package fsutil

import (
	"os"
)

// FileType represents a file, directory, or glob expression
type FileType int

// FileTypes
const (
	Unknown FileType = iota
	File
	Directory
	GlobExpression
)

// InferFileType takes a local path and tries to resolve it to a File, Directory
// or a glob expression
func InferFileType(path string) (FileType, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return GlobExpression, nil
		}
		return Unknown, err
	}
	if info.IsDir() {
		return Directory, nil
	}
	return File, nil
}
