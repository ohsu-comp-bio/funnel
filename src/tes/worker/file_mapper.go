package tesTaskengineWorker

import (
	"fmt"
	"golang.org/x/net/context"
	"io/ioutil"
	"os"
	"path"
	"tes/ga4gh"
	"tes/server/proto"
)

// FileMapper documentation
// TODO: documentation
type FileMapper struct {
	fileSystem FileSystemAccess
	VolumeDir  string
	client     *ga4gh_task_ref.SchedulerClient
	jobs       map[string]*JobFileMapper
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
	HostPath      string
	ContainerPath string
	Mode          string
}

// NewFileMapper documentation
// TODO: documentation
func NewFileMapper(client *ga4gh_task_ref.SchedulerClient, fileSystem FileSystemAccess, volumeDir string) *FileMapper {
	if _, err := os.Stat(volumeDir); os.IsNotExist(err) {
		os.Mkdir(volumeDir, 0700)
	}
	return &FileMapper{VolumeDir: volumeDir, jobs: make(map[string]*JobFileMapper), client: client, fileSystem: fileSystem}
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

// AddVolume documentation
// TODO: documentation
func (fileMapper *FileMapper) AddVolume(jobID string, source string, mount string) {
	tmpPath, _ := ioutil.TempDir(fileMapper.VolumeDir, fmt.Sprintf("job_%s", jobID))
	b := FSBinding{
		HostPath:      tmpPath,
		ContainerPath: mount,
		Mode:          "rw",
	}
	j := fileMapper.jobs[jobID]
	j.Bindings = append(j.Bindings, b)
}

// HostPath documentation
// TODO: documentation
func (fileMapper *FileMapper) HostPath(jobID string, mountPath string) string {
	for _, vol := range fileMapper.jobs[jobID].Bindings {
		base, relpath := pathMatch(vol.ContainerPath, mountPath)
		if len(base) > 0 {
			return path.Join(vol.HostPath, relpath)
		}
	}
	return ""
}

// MapInput documentation
// TODO: documentation
func (fileMapper *FileMapper) MapInput(jobID string, storage string, mountPath string, class string) error {
	for _, vol := range fileMapper.jobs[jobID].Bindings {
		base, relpath := pathMatch(vol.ContainerPath, mountPath)
		if len(base) > 0 {
			dstPath := path.Join(vol.HostPath, relpath)
			fmt.Printf("get %s %s\n", storage, dstPath)
			err := fileMapper.fileSystem.Get(storage, dstPath, class)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// MapOutput documentation
// TODO: documentation
func (fileMapper *FileMapper) MapOutput(jobID string, storage string, mountPath string, class string, create bool) error {
	a := ga4gh_task_exec.TaskParameter{Location: storage, Path: mountPath, Create: create, Class: class}
	j := fileMapper.jobs[jobID]
	if create {
		for _, vol := range fileMapper.jobs[jobID].Bindings {
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

// GetBindings documentation
// TODO: documentation
func (fileMapper *FileMapper) GetBindings(jobID string) []string {
	out := make([]string, 0, 10)
	for _, c := range fileMapper.jobs[jobID].Bindings {
		o := fmt.Sprintf("%s:%s:%s", c.HostPath, c.ContainerPath, c.Mode)
		out = append(out, o)
	}
	return out
}

// UpdateOutputs documentation
// TODO: documentation
func (fileMapper *FileMapper) UpdateOutputs(jobID string, jobNum int, exitCode int, stdoutText string, stderrText string) {
	log := ga4gh_task_exec.JobLog{Stdout: stdoutText, Stderr: stderrText, ExitCode: int32(exitCode)}
	a := ga4gh_task_ref.UpdateStatusRequest{Id: jobID, Step: int64(jobNum), Log: &log}
	(*fileMapper.client).UpdateJobStatus(context.Background(), &a)
}

// TempFile documentation
// TODO: documentation
func (fileMapper *FileMapper) TempFile(jobID string) (f *os.File, err error) {
	out, err := ioutil.TempFile(fileMapper.jobs[jobID].WorkDir, "ga4ghtask_")
	return out, err
}

// FinalizeJob documentation
// TODO: documentation
func (fileMapper *FileMapper) FinalizeJob(jobID string) {
	for _, out := range fileMapper.jobs[jobID].Outputs {
		hst := fileMapper.HostPath(jobID, out.Path)
		fileMapper.fileSystem.Put(out.Location, hst, out.Class)
	}
}
