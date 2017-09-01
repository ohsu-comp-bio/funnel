package compute

import (
	"fmt"
	"os"
)

// DetectFunnelBinaryPath detects the path to the "funnel" binary
func DetectFunnelBinaryPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("Failed to detect path of funnel binary")
	}
	return path, err
}
