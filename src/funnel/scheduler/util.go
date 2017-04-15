package scheduler

import (
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
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

// GenWorkerID returns a UUID string.
func GenWorkerID(prefix string) string {
	u, _ := uuid.NewV4()
	return fmt.Sprintf("%s-worker-%s", prefix, u.String())
}
