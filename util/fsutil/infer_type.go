package fsutil

import (
	"os"
)

type FileType int

const (
	File FileType = iota
	Directory
	GlobExpression
	Unknown
)

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
