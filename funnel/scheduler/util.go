package scheduler

import (
	"os"
	"path"
	"path/filepath"
)

// DetectWorkerPath detects the path to the "funnel" binary
func DetectWorkerPath() string {
	// TODO HACK: get the path to the worker executable
	//      move this to overrideable default config value?
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return path.Join(dir, "funnel")
}
