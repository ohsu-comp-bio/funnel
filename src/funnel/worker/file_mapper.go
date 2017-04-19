package worker

import (
	"fmt"
	"funnel/proto/tes"
	"funnel/util"
	proto "github.com/golang/protobuf/proto"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// FileMapper is responsible for mapping paths into a working directory on the
// worker's host file system.
//
// Every task needs it's own directory to work in. When a file is downloaded for
// a task, it needs to be stored in the task's working directory. Similar for task
// outputs, uploads, stdin/out/err, etc. FileMapper helps the worker engine
// manage all these paths.
type FileMapper struct {
	Volumes []Volume
	Inputs  []*tes.TaskParameter
	Outputs []*tes.TaskParameter
	dir     string
}

// Volume represents a volume mounted into a docker container.
// This includes a HostPath, the path on the host file system,
// and a ContainerPath, the path on the container file system,
// and whether the volume is read-only.
type Volume struct {
	// The path in tes worker.
	HostPath string
	// The path in Docker.
	ContainerPath string
	Readonly      bool
}

// NewFileMapper returns a new FileMapper, which maps files into the given
// base directory.
func NewFileMapper(dir string) *FileMapper {
	// TODO error handling
	dir, _ = filepath.Abs(dir)
	return &FileMapper{
		Volumes: []Volume{},
		Inputs:  []*tes.TaskParameter{},
		Outputs: []*tes.TaskParameter{},
		dir:     dir,
	}
}

// MapTask adds all the volumes, inputs, and outputs in the given Task to the FileMapper.
func (mapper *FileMapper) MapTask(task *tes.Task) error {

	// Add all the volumes to the mapper
	for _, vol := range task.Volumes {
		err := mapper.AddTmpVolume(vol)
		if err != nil {
			return err
		}
	}

	// Add all the inputs to the mapper
	for _, input := range task.Inputs {
		err := mapper.AddInput(input)
		if err != nil {
			return err
		}
	}

	// Add all the outputs to the mapper
	for _, output := range task.Outputs {
		err := mapper.AddOutput(output)
		if err != nil {
			return err
		}
	}

	return nil
}

// AddVolume adds a mapped volume to the mapper. A corresponding Volume record
// is added to mapper.Volumes.
//
// If the volume paths are invalid or can't be mapped, an error is returned.
func (mapper *FileMapper) AddVolume(hostPath string, mountPoint string, readonly bool) error {
	v := Volume{
		HostPath:      hostPath,
		ContainerPath: mountPoint,
		Readonly:      readonly,
	}

	if !mapper.hasVolume(v) {
		mapper.Volumes = append(mapper.Volumes, v)
	}

	return nil
}

// HostPath returns a mapped path.
//
// The path is concatenated to the mapper's base dir.
// e.g. If the mapper is configured with a base dir of "/tmp/mapped_files", then
// mapper.HostPath("/home/ubuntu/myfile") will return "/tmp/mapped_files/home/ubuntu/myfile".
//
// The mapped path is required to be a subpath of the mapper's base directory.
// e.g. mapper.HostPath("../../foo") should fail with an error.
func (mapper *FileMapper) HostPath(src string) (string, error) {
	p := path.Join(mapper.dir, src)
	p = path.Clean(p)
	if !mapper.IsSubpath(p, mapper.dir) {
		return "", fmt.Errorf("Invalid path: %s is not a valid subpath of %s", p, mapper.dir)
	}
	return p, nil
}

// OpenHostFile opens a file on the host file system at a mapped path.
// "src" is an unmapped path. This function will handle mapping the path.
//
// This function calls os.Open
//
// If the path can't be mapped or the file can't be opened, an error is returned.
func (mapper *FileMapper) OpenHostFile(src string) (*os.File, error) {
	p, perr := mapper.HostPath(src)
	if perr != nil {
		return nil, perr
	}
	f, oerr := os.Open(p)
	if oerr != nil {
		return nil, oerr
	}
	return f, nil
}

// CreateHostFile creates a file on the host file system at a mapped path.
// "src" is an unmapped path. This function will handle mapping the path.
//
// This function calls os.Create
//
// If the path can't be mapped or the file can't be created, an error is returned.
func (mapper *FileMapper) CreateHostFile(src string) (*os.File, error) {
	p, perr := mapper.HostPath(src)
	if perr != nil {
		return nil, perr
	}
	err := util.EnsurePath(p)
	if err != nil {
		return nil, err
	}
	f, oerr := os.Create(p)
	if oerr != nil {
		return nil, oerr
	}
	return f, nil
}

// AddTmpVolume creates a directory on the host based on the delcared path in
// the container and adds it to mapper.Volumes.
//
// If the path can't be mapped, an error is returned.
func (mapper *FileMapper) AddTmpVolume(mountPoint string) error {
	hostPath, err := mapper.HostPath(mountPoint)
	if err != nil {
		return err
	}

	err = util.EnsureDir(hostPath)
	if err != nil {
		return err
	}

	err = mapper.AddVolume(hostPath, mountPoint, false)
	if err != nil {
		return err
	}
	return nil
}

// AddInput adds an input to the mapped files for the given TaskParameter.
// A copy of the TaskParameter will be added to mapper.Inputs, with the
// "Path" field updated to the mapped host path.
//
// If the path can't be mapped an error is returned.
func (mapper *FileMapper) AddInput(input *tes.TaskParameter) error {
	hostPath, err := mapper.HostPath(input.Path)
	if err != nil {
		return err
	}

	err = util.EnsurePath(hostPath)
	if err != nil {
		return err
	}

	// Add input volumes
	err = mapper.AddVolume(hostPath, input.Path, true)
	if err != nil {
		return err
	}

	// Create a TaskParameter for the input with a path mapped to the host
	hostIn := proto.Clone(input).(*tes.TaskParameter)
	hostIn.Path = hostPath
	mapper.Inputs = append(mapper.Inputs, hostIn)
	return nil
}

// AddOutput adds an output to the mapped files for the given TaskParameter.
// A copy of the TaskParameter will be added to mapper.Outputs, with the
// "Path" field updated to the mapped host path.
//
// If the path can't be mapped, an error is returned.
func (mapper *FileMapper) AddOutput(output *tes.TaskParameter) error {
	hostPath, err := mapper.HostPath(output.Path)
	if err != nil {
		return err
	}

	hostDir := hostPath
	mountDir := output.Path
	if output.Type == tes.FileType_FILE {
		hostDir = path.Dir(hostPath)
		mountDir = path.Dir(output.Path)
	}

	err = util.EnsureDir(hostDir)
	if err != nil {
		return err
	}

	// Add output volumes
	err = mapper.AddVolume(hostDir, mountDir, false)
	if err != nil {
		return err
	}

	// Create a TaskParameter for the out with a path mapped to the host
	hostOut := proto.Clone(output).(*tes.TaskParameter)
	hostOut.Path = hostPath
	mapper.Outputs = append(mapper.Outputs, hostOut)
	return nil
}

// IsSubpath returns true if the given path "p" is a subpath of "base".
func (mapper *FileMapper) IsSubpath(p string, base string) bool {
	return strings.HasPrefix(p, base)
}

// check if a Volume is already present in the FileMapper
func (mapper *FileMapper) hasVolume(new Volume) bool {
	for _, v := range mapper.Volumes {
		if v == new {
			return true
		}
	}
	return false
}
