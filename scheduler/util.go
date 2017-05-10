package scheduler

import (
	"fmt"
	uuid "github.com/nu7hatch/gouuid"
	"os"
)

// DetectWorkerPath detects the path to the "funnel" binary
func DetectWorkerPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("Failed to detect path of funnel binary")
	}
	return path, err
}

// GenWorkerID returns a UUID string.
func GenWorkerID(prefix string) string {
	u, _ := uuid.NewV4()
	return fmt.Sprintf("%s-worker-%s", prefix, u.String())
}
