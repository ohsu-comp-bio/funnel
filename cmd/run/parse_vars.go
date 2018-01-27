package run

import (
	"errors"
	"fmt"
	"github.com/kballard/go-shellquote"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

// ErrKeyFmt describes an error in input/output/env/tag flag formatting
var ErrKeyFmt = errors.New("Arguments passed to --in, --out and --env must be of the form: KEY=VALUE")

// ErrStorageScheme describes an error in supported storage URL schemes.
var ErrStorageScheme = errors.New("File paths must be prefixed with one of:\n file://\n gs://\n s3://")

// DuplicateKeyErr returns a new error describing conflicting keys for env. vars., inputs, and outputs.
func DuplicateKeyErr(key string) error {
	return errors.New("Can't use the same KEY for multiple --in, --out, --env arguments: " + key)
}

// Parse CLI variable definitions (e.g "varname=value") into usable task values.
func valsToTask(vals flagVals) (task *tes.Task, err error) {

	// Any error occurring during parsing the variables an preparing the task
	// is a fatal error, so I'm using panic/recover to simplify error handling.
	defer func() {
		if x := recover(); x != nil {
			err = x.(error)
		}
	}()

	environ := map[string]string{}

	// Build the task message
	task = &tes.Task{
		Name:        vals.name,
		Description: vals.description,
		Resources: &tes.Resources{
			CpuCores:    uint32(vals.cpu),
			RamGb:       vals.ram,
			DiskGb:      vals.disk,
			Zones:       vals.zones,
			Preemptible: vals.preemptible,
		},
		Tags: map[string]string{},
    Image:   vals.container,
    Command: cmd,
    Workdir: vals.workdir,
    Env:     environ,
    Stdin:   stdin,
	}

  // Split command string based on shell syntax.
  cmd, _ := shellquote.Split(exec.cmd)
  stdinPath := fmt.Sprintf("/opt/funnel/inputs/stdin-%d", i)

  // Only set the stdin path if the --stdin flag was used.
  var stdin string
  if exec.stdin != "" {
    stdin = stdinPath
    task.Inputs = append(task.Inputs, &tes.Input{
      Name: fmt.Sprintf("stdin-%d", i),
      Url:  resolvePath(exec.stdin),
      Path: stdinPath,
    })
  }

	// Helper to make sure variable keys are unique.
	setenv := func(key, val string) {
		_, exists := environ[key]
		if exists {
			panic(DuplicateKeyErr(key))
		}
		environ[key] = val
	}

	for _, raw := range vals.inputs {
		k, v := parseCliVar(raw)
		url := resolvePath(v)
		path := "/opt/funnel/inputs/" + stripStoragePrefix(url)
		setenv(k, path)
		task.Inputs = append(task.Inputs, &tes.Input{
			Name: k,
			Url:  url,
			Path: path,
		})
	}

	for _, raw := range vals.inputDirs {
		k, v := parseCliVar(raw)
		url := resolvePath(v)
		path := "/opt/funnel/inputs/" + stripStoragePrefix(url)
		setenv(k, path)
		task.Inputs = append(task.Inputs, &tes.Input{
			Name: k,
			Url:  url,
			Path: path,
			Type: tes.FileType_DIRECTORY,
		})
	}

	for _, raw := range vals.content {
		k, v := parseCliVar(raw)
		path := "/opt/funnel/inputs/" + stripStoragePrefix(resolvePath(v))
		setenv(k, path)
		task.Inputs = append(task.Inputs, &tes.Input{
			Name:    k,
			Path:    path,
			Content: getContent(v),
		})
	}

	for _, raw := range vals.outputs {
		k, v := parseCliVar(raw)
		url := resolvePath(v)
		path := "/opt/funnel/outputs/" + stripStoragePrefix(url)
		setenv(k, path)
		task.Outputs = append(task.Outputs, &tes.Output{
			Name: k,
			Url:  url,
			Path: path,
		})
	}

	for _, raw := range vals.outputDirs {
		k, v := parseCliVar(raw)
		url := resolvePath(v)
		path := "/opt/funnel/outputs/" + stripStoragePrefix(url)
		setenv(k, path)
		task.Outputs = append(task.Outputs, &tes.Output{
			Name: k,
			Url:  url,
			Path: path,
			Type: tes.FileType_DIRECTORY,
		})
	}

	for _, raw := range vals.environ {
		k, v := parseCliVar(raw)
		setenv(k, v)
	}

	for _, raw := range vals.tags {
		k, v := parseCliVar(raw)
		task.Tags[k] = v
	}
	return
}

func getContent(p string) string {
	b, err := ioutil.ReadFile(p)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func parseCliVar(raw string) (string, string) {
	re := regexp.MustCompile("=")
	res := re.Split(raw, -1)

	if len(res) != 2 {
		panic(ErrKeyFmt)
	}

	key := res[0]
	val := res[1]
	return key, val
}

// Give a input/output URL "raw", return the path of the file
// relative to the container.
func containerPath(raw, base string) string {
	url := resolvePath(raw)
	p := stripStoragePrefix(url)
	return base + p
}

func stripStoragePrefix(url string) string {
	re := regexp.MustCompile("[a-z0-9]+://")
	if !re.MatchString(url) {
		panic(ErrStorageScheme)
	}
	path := re.ReplaceAllString(url, "")
	return strings.TrimPrefix(path, "/")
}

func resolvePath(url string) string {
	re := regexp.MustCompile("[a-z0-9]+://")
	prefixed := re.MatchString(url)
	local := strings.HasPrefix(url, "/") || strings.HasPrefix(url, ".") ||
		strings.HasPrefix(url, "~") || !prefixed

	switch {
	case local:
		path, err := filepath.Abs(url)
		if err != nil {
			panic(err)
		}
		return "file://" + path
	case prefixed:
		return url
	default:
		panic(fmt.Errorf("could not resolve filepath: %s", url))
	}
}
