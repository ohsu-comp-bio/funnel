package compute

import (
	"fmt"
	"os"
)

// detectFunnelBinaryPath detects the path to the "funnel" binary
func detectFunnelBinaryPath() (string, error) {
	path, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("Failed to detect path of funnel binary")
	}
	return path, err
}
