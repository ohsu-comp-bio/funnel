package tesTaskEngineWorker

import (
	"fmt"
	"io"
	"os"
	"path"
)

func pathMatch(base string, query string) (string, string) {
	if path.Clean(base) == path.Clean(query) {
		return query, ""
	}
	dir, file := path.Split(query)
	if len(dir) > 1 {
		d, p := pathMatch(base, dir)
		return d, path.Join(p, file)
	}
	return "", ""
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

// CopyFile documentation
// TODO: Documentation
func CopyFile(src, dst string) (err error) {
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

// CopyDir documentation
// TODO: Documentation
func CopyDir(source string, dest string) (err error) {
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
			err = CopyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// Performs copy.
			err = CopyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}
