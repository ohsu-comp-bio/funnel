

package ga4gh_taskengine

import (
	"os"
	"ga4gh-tasks"
	"fmt"
	"io/ioutil"
	"path"
	"ga4gh-server/proto"
	"golang.org/x/net/context"
)


type FileMapper interface {
	Job(jobId string)
	MapInput(jobId string, storagePath string, localPath string, directory bool) error
	MapOutput(jobId string, storagePath string, localPath string, directory bool) error

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
	HostPath string
	ContainerPath string
	Mode string
}


func NewSharedFS(client *ga4gh_task_ref.SchedulerClient, storageDir string, diskDir string) *SharedFileMapper {
	if _, err := os.Stat(storageDir); os.IsNotExist(err) {
		os.Mkdir(storageDir, 0700)
	}
	if _, err := os.Stat(diskDir); os.IsNotExist(err) {
		os.Mkdir(diskDir, 0700)
	}

	return &SharedFileMapper{StorageDir: storageDir, DiskDir: diskDir, jobs: make(map[string]JobSharedFileMapper), client:client}
}

type JobSharedFileMapper struct {
	JobId string
	WorkDir string
	Bindings []FSBinding
}

type SharedFileMapper struct {
	StorageDir string
	DiskDir string
	client *ga4gh_task_ref.SchedulerClient
	jobs map[string]JobSharedFileMapper
}

func (self *SharedFileMapper) Job(jobId string) {
	//create a working 'disk' for runtime files
	w := path.Join(self.DiskDir, jobId)
	if _, err := os.Stat(w); err != nil {
		os.Mkdir(w, 0700)
	}
	a := JobSharedFileMapper{JobId:jobId, WorkDir:w}
	self.jobs[jobId] = a
}

func (self *SharedFileMapper) MapInput(jobId string, storage string, mountPath string, directory bool) error {
	//because we're running on a shared file system, there is no need to copy
	//the 'storage' path to local disk, we can just mount it directly
	srcPath := path.Join(self.StorageDir, storage)
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		return fmt.Errorf("storage file '%s' not found", srcPath)
	}
	b := FSBinding {
		HostPath: srcPath,
		ContainerPath: mountPath,
		Mode: "rw",
	}
	j := self.jobs[jobId]
	j.Bindings = append(j.Bindings, b)
	return nil
}

func (self *SharedFileMapper) MapOutput(jobId string, storage string, mountPath string, directory bool) error {
	var diskDir string

	//diskDir = path.Join(self.DiskDir, disk)
	//if (disk == "") {
		diskDir = self.jobs[jobId].WorkDir
	//}
	var hostPath string
	if (directory) {
		hostPath, _ = ioutil.TempDir(diskDir, "outdir_" )
	} else {
		hostFile, _ := ioutil.TempFile(diskDir, "outfile_")
		hostPath = hostFile.Name()
		hostFile.Close()
	}

	b := FSBinding {
		HostPath: hostPath,
		ContainerPath: mountPath,
		Mode: "rw",
	}
	j := self.jobs[jobId]
	j.Bindings = append(j.Bindings, b)
	return nil
}


func (self *SharedFileMapper) GetBindings(jobId string) []string {
	out := make([]string, 0, 10)
	for _, c := range self.jobs[jobId].Bindings {
		o := fmt.Sprint("%s:%s:%s", c.HostPath, c.ContainerPath, c.Mode)
		out = append(out, o)
	}
	return out
}


func (self *SharedFileMapper) UpdateOutputs(jobId string, jobNum int, exitCode int, stdoutText string, stderrText string) {
	log := ga4gh_task_exec.JobLog{Stdout:stdoutText, Stderr:stderrText, ExitCode:int32(exitCode)}
	a := ga4gh_task_ref.UpdateStatusRequest{Id:jobId, Step:int64(jobNum), Log:&log }
	(*self.client).UpdateJobStatus(context.Background(), &a)
}


func (self *SharedFileMapper) TempFile(jobId string) (f *os.File, err error) {
	out, err := ioutil.TempFile(self.jobs[jobId].WorkDir, "ga4ghtask_")
	return out, err
}



func (self *SharedFileMapper) FinalizeJob(jobId string) {
	//nothing to do, the files are already in their place
}

