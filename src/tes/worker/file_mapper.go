package worker

import (
	"fmt"
	proto "github.com/golang/protobuf/proto"
	"os"
	"path"
	"path/filepath"
	"strings"
	pbe "tes/ga4gh"
)

// FileMapper is responsible for mapping paths into a working directory on the
// worker's host file system.
//
// Every job needs it's own directory to work in. When a file is downloaded for
// a job, it needs to be stored in the job's working directory. Similar for job
// outputs, uploads, stdin/out/err, etc. FileMapper helps the worker engine
// manage all these paths.
type FileMapper struct {
	Volumes []Volume
	Inputs  []*pbe.TaskParameter
	Outputs []*pbe.TaskParameter
	dir     string
}

// Volume represents a volume mounted into a docker container.
// This includes a HostPath, the path on the host file system,
// and a ContainerPath, the path on the container file system,
// and the mode ("rw" = read-only, "ro" = read-write).
type Volume struct {
	// The path in tes worker.
	HostPath string
	// The path in Docker.
	ContainerPath string
	Mode          string
}

// NewJobFileMapper returns a new FileMapper configured to map files for a job.
//
// The following example will return a FileMapper that maps into the
// "/path/to/workdir/123/" directory on the host file system.
//     NewJobFileMapper("123", "/path/to/workdir")
func NewJobFileMapper(jobID string, baseDir string) *FileMapper {
	dir := path.Join(baseDir, jobID)
	// TODO error handling
	dir, _ = filepath.Abs(dir)
	return &FileMapper{dir: dir}
}

// getMapper returns a FileMapper instance with volumes, inputs, and outputs
// configured for the given job.
func (mapper *FileMapper) MapTask(task *pbe.Task) error {

	// Add all the volumes to the mapper
	for _, vol := range task.Resources.Volumes {
		err := mapper.AddVolume(vol.Source, vol.MountPoint)
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
// Currently, volumes are hard-coded to "rw" (read-write).
//
// If the volume paths are invalid or can't be mapped, an error is returned.
func (mapper *FileMapper) AddVolume(source string, mountPoint string) error {
	if source != "" {
		return fmt.Errorf("Could not create a volume: 'source' is not supported for %s", source)
	}
	if mountPoint == "" {
		return fmt.Errorf("Could not create a volume: 'mountPoint' is required for %s", mountPoint)
	}

	hostPath, err := mapper.HostPath(mountPoint)
	if err != nil {
		return err
	}

	v := Volume{
		HostPath:      hostPath,
		ContainerPath: mountPoint,
		// TODO should be read only?
		Mode: "rw",
	}

	// Ensure that the volume directory exists on the host
	perr := ensureDir(hostPath)
	if perr != nil {
		return perr
	}

	mapper.Volumes = append(mapper.Volumes, v)
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
	err := ensurePath(p)
	if err != nil {
		return nil, err
	}
	f, oerr := os.Create(p)
	if oerr != nil {
		return nil, oerr
	}
	return f, nil
}

// AddInput adds an input to the mapped files for the given TaskParameter.
// A copy of the TaskParameter will be added to mapper.Inputs, with the
// "Path" field updated to the mapped host path.
//
// If the path can't be mapped, or the path is not in an existing volume,
// an error is returned.
func (mapper *FileMapper) AddInput(input *pbe.TaskParameter) error {
	p, err := mapper.HostPath(input.Path)
	if err != nil {
		return err
	}
	// Require that the path be in a defined volume
	if !mapper.IsInVolume(p) {
		return fmt.Errorf("Input path is required to be in a volume: %s", input.Path)
	}

	perr := ensurePath(p)
	if perr != nil {
		return perr
	}

	// Create a TaskParameter for the input with a path mapped to the host
	hostIn := proto.Clone(input).(*pbe.TaskParameter)
	hostIn.Path = p
	mapper.Inputs = append(mapper.Inputs, hostIn)
	return nil
}

// AddOutput adds an output to the mapped files for the given TaskParameter.
// A copy of the TaskParameter will be added to mapper.Outputs, with the
// "Path" field updated to the mapped host path.
//
// If the Create flag is set on the TaskParameter, the file will be created
// on the host file system.
//
// If the path can't be mapped, or the path is not in an existing volume,
// an error is returned.
func (mapper *FileMapper) AddOutput(output *pbe.TaskParameter) error {
	p, err := mapper.HostPath(output.Path)
	if err != nil {
		return err
	}
	// Require that the path be in a defined volume
	if !mapper.IsInVolume(p) {
		return fmt.Errorf("Output path is required to be in a volume: %s", output.Path)
	}
	// Create the file if needed, as per the TES spec
	if output.Create {
		err := ensureFile(p, output.Class)
		if err != nil {
			return err
		}
	}
	// Create a TaskParameter for the out with a path mapped to the host
	hostOut := proto.Clone(output).(*pbe.TaskParameter)
	hostOut.Path = p
	mapper.Outputs = append(mapper.Outputs, hostOut)
	return nil
}

// IsSubpath returns true if the given path "p" is a subpath of "base".
func (mapper *FileMapper) IsSubpath(p string, base string) bool {
	return strings.HasPrefix(p, base)
}

// IsInVolume checks whether a given path is in a mapped volume.
func (mapper *FileMapper) IsInVolume(p string) bool {
	for _, vol := range mapper.Volumes {
		if mapper.IsSubpath(p, vol.HostPath) {
			return true
		}
	}
	return false
}
