package tesTaskEngineWorker

import (
	"os"
	"path"
)

const headerSize = int64(102400)

// exists returns whether the given file or directory exists or not
func exists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func ensureDir(p string) error {
	e, err := exists(p)
	if err != nil {
		return err
	}
	if !e {
		// TODO configurable mode?
		err := os.MkdirAll(p, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

func ensurePath(p string) error {
	dir := path.Dir(p)
	return ensureDir(dir)
}

func ensureFile(p string, class string) error {
	err := ensurePath(p)
	if err != nil {
		return err
	}
	if class == "File" {
		f, err := os.Create(p)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

func readFileHead(path string) string {
	// TODO handle errors?
	f, _ := os.Open(path)
	buffer := make([]byte, headerSize)
	l, _ := f.Read(buffer)
	f.Close()
	return string(buffer[:l])
}
