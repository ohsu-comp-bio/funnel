package tesTaskEngineWorker

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"tes/ga4gh"
	"tes/server/proto"
)

// FileMapper documentation
// TODO: documentation
type FileMapper struct {
	fileSystems map[string]FileSystemAccess
	VolumeDir   string
	client      *ga4gh_task_ref.SchedulerClient
	jobs        map[string]*JobFileMapper
}

// JobFileMapper documentation
// TODO: documentation
type JobFileMapper struct {
	JobID    string
	WorkDir  string
	Bindings []FSBinding
	Outputs  []ga4gh_task_exec.TaskParameter
}

// FileSystemAccess documentation
// TODO: documentation
type FileSystemAccess interface {
	Get(storage string, path string, class string) error
	Put(storage string, path string, class string) error
}

// EngineStatus documentation
// TODO: documentation
type EngineStatus struct {
	JobCount   int32
	ActiveJobs int32
}

// FSBinding documentation
// TODO: documentation
type FSBinding struct {
	// The path in tes worker.
	HostPath string
	// The path in Docker.
	ContainerPath string
	Mode          string
}

// NewFileMapper documentation
// TODO: documentation

func NewFileMapper(fileSystem FileSystemAccess, volumeDir string) *FileMapper {
	if _, err := os.Stat(volumeDir); os.IsNotExist(err) {
		os.Mkdir(volumeDir, 0700)
	}
	return &FileMapper{VolumeDir: volumeDir, jobs: make(map[string]*JobFileMapper)}
}

// Job documentation
// TODO: documentation
func (fileMapper *FileMapper) Job(jobID string) {
	//create a working 'disk' for runtime files
	w := path.Join(fileMapper.VolumeDir, jobID)
	if _, err := os.Stat(w); err != nil {
		os.Mkdir(w, 0700)
	}
	a := JobFileMapper{JobID: jobID, WorkDir: w}
	fileMapper.jobs[jobID] = &a
}

// AddVolume makes the given `mount` to be the ContainerPath.
func (fileMapper *FileMapper) AddVolume(jobID string, source string, mount string) {
	// The reason we use `tmpPath` is that we're pulling the file
	// from the object store to a temporary path, which will be
	// used as a working directory. We don't need to store the
	// intermediate files, so we use a temporary path.
	tmpPath, _ := ioutil.TempDir(fileMapper.VolumeDir, fmt.Sprintf("job_%s", jobID))
	b := FSBinding{
		HostPath:      tmpPath,
		ContainerPath: mount,
		Mode:          "rw",
	}
	j := fileMapper.jobs[jobID]
	j.Bindings = append(j.Bindings, b)
}

// HostPath returns a path from the `HostPath` that is the
// equidistance from ContainerPath to mountPath.
func (fileMapper *FileMapper) HostPath(jobID string, mountPath string) string {
	for _, vol := range fileMapper.jobs[jobID].Bindings {
		base, relpath := pathMatch(vol.ContainerPath, mountPath)
		if len(base) > 0 {
			return path.Join(vol.HostPath, relpath)
		}
	}
	return ""
}

func (fileMapper *FileMapper) FindFS(url string) (FileSystemAccess, error) {
	tmp := strings.Split(url, ":")[0]
	fs, ok := fileMapper.fileSystems[tmp]
	if !ok {
		return fs, fmt.Errorf("File System %s not found", tmp)
	}
	return fs, nil
}

// MapInput gets the file and put it into fileMapper. `storage` is
// related to swift object store.
func (fileMapper *FileMapper) MapInput(jobID string, storage string, mountPath string, class string) error {
	for _, vol := range fileMapper.jobs[jobID].Bindings {
		// Finds the relative path to mountPath.
		base, relpath := pathMatch(vol.ContainerPath, mountPath)
		if len(base) > 0 {
			// HostPath is a tmpPath.
			// dst{ath is destination path.
			dstPath := path.Join(vol.HostPath, relpath)
			fmt.Printf("get %s %s\n", storage, dstPath)
			fs, err := fileMapper.FindFS(storage)
			if err != nil {
				return err
			}
			// Copies storage to dstPath.  While the
			// result is stored in `err` if `err` is nil,
			// this operation did not throw an error.
			err = fs.Get(storage, dstPath, class)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// MapOutput adds the output directory.
func (fileMapper *FileMapper) MapOutput(jobID string, storage string, mountPath string, class string, create bool) error {
	a := ga4gh_task_exec.TaskParameter{Location: storage, Path: mountPath, Create: create, Class: class}
	j := fileMapper.jobs[jobID]
	// If create is True, make a directory under the path if class
	// is "Directory" or make a file under the path if class is
	// "File".
	if create {
		// Iterate through fileMapper.jobs.Bindings returns
		// volumes, which are file system bindings.
		for _, vol := range fileMapper.jobs[jobID].Bindings {
			// If this path is in docker
			// (vol.ContainerPath), add the output
			// directory to j.Outputs.
			base, relpath := pathMatch(vol.ContainerPath, mountPath)
			if len(base) > 0 {
				if class == "Directory" {
					os.MkdirAll(path.Join(vol.HostPath, relpath), 0777)
				} else if class == "File" {
					ioutil.WriteFile(path.Join(vol.HostPath, relpath), []byte{}, 0777)
				} else {
					return fmt.Errorf("Unknown class type: %s", class)
				}
			}
		}
	}
	j.Outputs = append(j.Outputs, a)
	return nil
}

// GetBindings takes a jobID and returns an array of string.
func (fileMapper *FileMapper) GetBindings(jobID string) []string {
	// Makes a slice of string where len is 0, and capacity is 10.
	out := make([]string, 0, 10)
	// Goes through each binding. `c` is the binding, and we ignore the index.
	for _, c := range fileMapper.jobs[jobID].Bindings {
		// `out` is an argument for docker run later.
		// `o` is structed as "HostPath:ContainerPath:Mode".
		o := fmt.Sprintf("%s:%s:%s", c.HostPath, c.ContainerPath, c.Mode)
		out = append(out, o)
	}
	return out
}

// TempFile creates a temporary file and returns a pointer to an operating system file.
func (fileMapper *FileMapper) TempFile(jobID string) (f *os.File, err error) {
	out, err := ioutil.TempFile(fileMapper.jobs[jobID].WorkDir, "ga4ghtask_")
	return out, err
}

// FinalizeJob documentation
// TODO: documentation
func (fileMapper *FileMapper) FinalizeJob(jobID string) error {
	for _, out := range fileMapper.jobs[jobID].Outputs {
		hst := fileMapper.HostPath(jobID, out.Path)
		fs, err := fileMapper.FindFS(out.Location)
		if err != nil {
			return err
		}
		fs.Put(out.Location, hst, out.Class)
	}
	return nil
}
