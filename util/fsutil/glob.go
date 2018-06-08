package fsutil

import (
	"os"
	"path/filepath"

	"github.com/mattn/go-zglob"
)

func adjustRel(newroot string, files []Hostfile) []Hostfile {
	out := []Hostfile{}
	for _, f := range files {
		f.Rel = filepath.Join(newroot, f.Rel)
		out = append(out, f)
	}
	return out
}

// Glob returns a list of HostFile objects matching pattern or nil if there is no matching file.
// This method uses the implementation in github.com/mattn/go-zglob.
func Glob(pattern string) ([]Hostfile, error) {
	var files []Hostfile
	matches, err := zglob.Glob(pattern)
	if err != nil {
		return nil, err
	}
	for _, path := range matches {
		finfo, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if finfo.IsDir() {
			dfiles, err := WalkFiles(path)
			if err != nil {
				return nil, err
			}
			dfiles = adjustRel(path, dfiles)
			files = append(files, dfiles...)
			continue
		}
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		files = append(files, Hostfile{
			Rel:          path,
			Abs:          absPath,
			Size:         finfo.Size(),
			LastModified: finfo.ModTime(),
		})
	}
	return files, nil
}
