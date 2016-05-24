

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
	MapInput(jobId string, srcPath string, localCopy ga4gh_task_exec.LocalCopy)
	MapOutput(jobId string, localCopy ga4gh_task_exec.LocalCopy, dstPath string)

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


func NewSharedFS(client *ga4gh_task_ref.SchedulerClient,  workdir string) *SharedFileMapper {
	if _, err := os.Stat(workdir); os.IsNotExist(err) {
		os.Mkdir(workdir, 0700)
	}
	return &SharedFileMapper{WorkDir: workdir, jobs: make(map[string]JobSharedFileMapper), client:client}
}

type JobSharedFileMapper struct {
	JobId string
	WorkDir string
	Bindings []FSBinding
}

type SharedFileMapper struct {
	WorkDir string
	client *ga4gh_task_ref.SchedulerClient
	jobs map[string]JobSharedFileMapper
}

func (self *SharedFileMapper) Job(jobId string) {
	w := path.Join(self.WorkDir, jobId)
	if _, err := os.Stat(w); err != nil {
		os.Mkdir(w, 0700)
	}
	a := JobSharedFileMapper{JobId:jobId, WorkDir:w}
	self.jobs[jobId] = a
}

func (self *SharedFileMapper) MapInput(jobId string, srcPath string, localCopy ga4gh_task_exec.LocalCopy) {
	b := FSBinding {
		HostPath: srcPath,
		ContainerPath: localCopy.Path,
		Mode: "rw",
	}
	j := self.jobs[jobId]
	j.Bindings = append(j.Bindings, b)
}

func (self *SharedFileMapper) MapOutput(jobId string, localCopy ga4gh_task_exec.LocalCopy, dstPath string) {
	//nothing here yet
	b := FSBinding {
		HostPath: dstPath,
		ContainerPath: localCopy.Path,
		Mode: "rw",
	}
	j := self.jobs[jobId]
	j.Bindings = append(j.Bindings, b)
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
	log := ga4gh_task_exec.TaskOpLog{Stdout:stdoutText, Stderr:stderrText, ExitCode:int32(exitCode)}
	a := ga4gh_task_ref.UpdateStatusRequest{Id:jobId, Step:int64(jobNum), Log:&log }
	(*self.client).UpdateTaskOpStatus(context.Background(), &a)
}


func (self *SharedFileMapper) TempFile(jobId string) (f *os.File, err error) {
	out, err := ioutil.TempFile(self.jobs[jobId].WorkDir, "ga4ghtask_")
	return out, err
}



func (self *SharedFileMapper) FinalizeJob(jobId string) {
	//nothing to do, the files are already in their place
}

