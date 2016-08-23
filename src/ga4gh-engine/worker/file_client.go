package ga4gh_taskengine_worker

import (
	"fmt"
	"ga4gh-server/proto"
	"ga4gh-tasks"
	"golang.org/x/net/context"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

type FileMapper interface {
	Job(jobId string)
	AddVolume(jobId string, source string, mount string)
	MapInput(jobId string, storagePath string, localPath string, directory bool) error
	MapOutput(jobId string, storagePath string, localPath string, directory bool, create bool) error

	HostPath(jobId string, mountPath string) string

	TempFile(jobId string) (f *os.File, err error)
	GetBindings(jobId string) []string
	UpdateOutputs(jobId string, stepNum int, exit_code int, stdoutText string, stderrText string)

	FinalizeJob(jobId string)
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

func NewSharedFS(client *ga4gh_task_ref.SchedulerClient, storageDir string, volumeDir string) *SharedFileMapper {
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		os.Mkdir(storageDir, 0700)
	}
	if _, err := os.Stat(volumeDir); os.IsNotExist(err) {
		os.Mkdir(volumeDir, 0700)
	}

	return &SharedFileMapper{StorageDir: storageDir, VolumeDir: volumeDir, jobs: make(map[string]*JobSharedFileMapper), client: client}
}

type JobSharedFileMapper struct {
	JobId    string
	WorkDir  string
	Bindings []FSBinding
	Outputs  []ga4gh_task_exec.TaskParameter
}

type SharedFileMapper struct {
	StorageDir string
	VolumeDir  string
	client     *ga4gh_task_ref.SchedulerClient
	jobs       map[string]*JobSharedFileMapper
}

func (self *SharedFileMapper) Job(jobId string) {
	//create a working 'disk' for runtime files
	w := path.Join(self.VolumeDir, jobId)
	if _, err := os.Stat(w); err != nil {
		os.Mkdir(w, 0700)
	}
	a := JobSharedFileMapper{JobId: jobId, WorkDir: w}
	self.jobs[jobId] = &a
}

func (self *SharedFileMapper) AddVolume(jobId string, source string, mount string) {
	tmpPath, _ := ioutil.TempDir(self.VolumeDir, fmt.Sprintf("job_%s", jobId))
	b := FSBinding{
		HostPath:      tmpPath,
		ContainerPath: mount,
		Mode:          "rw",
	}
	j := self.jobs[jobId]
	j.Bindings = append(j.Bindings, b)
}

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

func (self *SharedFileMapper) HostPath(jobId string, mountPath string) string {
	for _, vol := range self.jobs[jobId].Bindings {
		base, relpath := pathMatch(vol.ContainerPath, mountPath)
		if len(base) > 0 {
			return path.Join(vol.HostPath, relpath)
		}
	}
	return ""
}

func (self *SharedFileMapper) MapInput(jobId string, storage string, mountPath string, directory bool) error {
	storage = strings.TrimPrefix(storage, "fs://")
	srcPath := path.Join(self.StorageDir, storage)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("storage file '%s' not found", srcPath)
	}

	for _, vol := range self.jobs[jobId].Bindings {
		base, relpath := pathMatch(vol.ContainerPath, mountPath)
		if len(base) > 0 {
			fmt.Printf("cp %s %s\n", srcPath, path.Join(vol.HostPath, relpath))
			copyFileContents(srcPath, path.Join(vol.HostPath, relpath))
		}
	}
	return nil
}

func (self *SharedFileMapper) MapOutput(jobId string, storage string, mountPath string, directory bool, create bool) error {
	a := ga4gh_task_exec.TaskParameter{Location: storage, Path: mountPath, Create: create, Directory: directory}
	j := self.jobs[jobId]
	if create {
		for _, vol := range self.jobs[jobId].Bindings {
			base, relpath := pathMatch(vol.ContainerPath, mountPath)
			if len(base) > 0 {
				if directory {
					os.MkdirAll(path.Join(vol.HostPath, relpath), 0777)
				} else {
					ioutil.WriteFile(path.Join(vol.HostPath, relpath), []byte{}, 0777)
				}
			}
		}
	}
	j.Outputs = append(j.Outputs, a)
	return nil
}

func (self *SharedFileMapper) GetBindings(jobId string) []string {
	out := make([]string, 0, 10)
	for _, c := range self.jobs[jobId].Bindings {
		o := fmt.Sprintf("%s:%s:%s", c.HostPath, c.ContainerPath, c.Mode)
		out = append(out, o)
	}
	return out
}

func (self *SharedFileMapper) UpdateOutputs(jobId string, jobNum int, exitCode int, stdoutText string, stderrText string) {
	log := ga4gh_task_exec.JobLog{Stdout: stdoutText, Stderr: stderrText, ExitCode: int32(exitCode)}
	a := ga4gh_task_ref.UpdateStatusRequest{Id: jobId, Step: int64(jobNum), Log: &log}
	(*self.client).UpdateJobStatus(context.Background(), &a)
}

func (self *SharedFileMapper) TempFile(jobId string) (f *os.File, err error) {
	out, err := ioutil.TempFile(self.jobs[jobId].WorkDir, "ga4ghtask_")
	return out, err
}

func (self *SharedFileMapper) FinalizeJob(jobId string) {
	for _, out := range self.jobs[jobId].Outputs {
		hst := self.HostPath(jobId, out.Path)
		storage := strings.TrimPrefix(out.Location, "fs://")
		log.Printf("copy out %s %s (%#v)\n", hst, path.Join(self.StorageDir, storage), out)
		//copy to storage directory
		if out.Directory {
			err := CopyDir(hst, path.Join(self.StorageDir, storage))
			if err != nil {
				log.Printf("Error copying output directory %s to %s", hst, out.Location)
			}
		} else {
			err := CopyFile(hst, path.Join(self.StorageDir, storage))
			if err != nil {
				log.Printf("Error copying output file %s to %s", hst, out.Location)
			}
		}
	}
}
