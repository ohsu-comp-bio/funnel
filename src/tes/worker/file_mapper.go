package tes_taskengine_worker

import (
	"fmt"
	"golang.org/x/net/context"
	"io/ioutil"
	"os"
	"path"
	"tes/ga4gh"
	"tes/server/proto"
)

type FileMapper struct {
	fileSystem FileSystemAccess
	VolumeDir  string
	client     *ga4gh_task_ref.SchedulerClient
	jobs       map[string]*JobFileMapper
}

type JobFileMapper struct {
	JobId    string
	WorkDir  string
	Bindings []FSBinding
	Outputs  []ga4gh_task_exec.TaskParameter
}

type FileSystemAccess interface {
	Get(storage string, path string, class string) error
	Put(storage string, path string, class string) error
}

type EngineStatus struct {
	JobCount   int32
	ActiveJobs int32
}

type FSBinding struct {
	HostPath      string
	ContainerPath string
	Mode          string
}

func NewFileMapper(client *ga4gh_task_ref.SchedulerClient, fileSystem FileSystemAccess, volumeDir string) *FileMapper {
	if _, err := os.Stat(volumeDir); os.IsNotExist(err) {
		os.Mkdir(volumeDir, 0700)
	}
	return &FileMapper{VolumeDir: volumeDir, jobs: make(map[string]*JobFileMapper), client: client, fileSystem: fileSystem}
}

func (self *FileMapper) Job(jobId string) {
	//create a working 'disk' for runtime files
	w := path.Join(self.VolumeDir, jobId)
	if _, err := os.Stat(w); err != nil {
		os.Mkdir(w, 0700)
	}
	a := JobFileMapper{JobId: jobId, WorkDir: w}
	self.jobs[jobId] = &a
}

func (self *FileMapper) AddVolume(jobId string, source string, mount string) {
	tmpPath, _ := ioutil.TempDir(self.VolumeDir, fmt.Sprintf("job_%s", jobId))
	b := FSBinding{
		HostPath:      tmpPath,
		ContainerPath: mount,
		Mode:          "rw",
	}
	j := self.jobs[jobId]
	j.Bindings = append(j.Bindings, b)
}

func (self *FileMapper) HostPath(jobId string, mountPath string) string {
	for _, vol := range self.jobs[jobId].Bindings {
		base, relpath := pathMatch(vol.ContainerPath, mountPath)
		if len(base) > 0 {
			return path.Join(vol.HostPath, relpath)
		}
	}
	return ""
}

func (self *FileMapper) MapInput(jobId string, storage string, mountPath string, class string) error {
	for _, vol := range self.jobs[jobId].Bindings {
		base, relpath := pathMatch(vol.ContainerPath, mountPath)
		if len(base) > 0 {
			dstPath := path.Join(vol.HostPath, relpath)
			fmt.Printf("get %s %s\n", storage, dstPath)
			err := self.fileSystem.Get(storage, dstPath, class)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *FileMapper) MapOutput(jobId string, storage string, mountPath string, class string, create bool) error {
	a := ga4gh_task_exec.TaskParameter{Location: storage, Path: mountPath, Create: create, Class: class}
	j := self.jobs[jobId]
	if create {
		for _, vol := range self.jobs[jobId].Bindings {
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

func (self *FileMapper) GetBindings(jobId string) []string {
	out := make([]string, 0, 10)
	for _, c := range self.jobs[jobId].Bindings {
		o := fmt.Sprintf("%s:%s:%s", c.HostPath, c.ContainerPath, c.Mode)
		out = append(out, o)
	}
	return out
}

func (self *FileMapper) UpdateOutputs(jobId string, jobNum int, exitCode int, stdoutText string, stderrText string) {
	log := ga4gh_task_exec.JobLog{Stdout: stdoutText, Stderr: stderrText, ExitCode: int32(exitCode)}
	a := ga4gh_task_ref.UpdateStatusRequest{Id: jobId, Step: int64(jobNum), Log: &log}
	(*self.client).UpdateJobStatus(context.Background(), &a)
}

func (self *FileMapper) TempFile(jobId string) (f *os.File, err error) {
	out, err := ioutil.TempFile(self.jobs[jobId].WorkDir, "ga4ghtask_")
	return out, err
}

func (self *FileMapper) FinalizeJob(jobId string) {
	for _, out := range self.jobs[jobId].Outputs {
		hst := self.HostPath(jobId, out.Path)
		self.fileSystem.Put(out.Location, hst, out.Class)
	}
}
