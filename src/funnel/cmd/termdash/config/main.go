// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package config

import (
	"fmt"
	"os"
)

var (
	GlobalParams   []*Param
	GlobalSwitches []*Switch
)

func init() {
	for _, p := range params {
		GlobalParams = append(GlobalParams, p)
	}
	for _, s := range switches {
		GlobalSwitches = append(GlobalSwitches, s)
	}
}

func quote(s string) string {
	return fmt.Sprintf("\"%s\"", s)
}

// Return env var value if set, else return defaultVal
func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	return defaultVal
}
