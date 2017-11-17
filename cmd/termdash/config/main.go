// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package config

var (
	GlobalParams   []*Param
	GlobalSwitches []*Switch
)

func Init() {
	for _, p := range params {
		GlobalParams = append(GlobalParams, p)
	}
	for _, s := range switches {
		GlobalSwitches = append(GlobalSwitches, s)
	}
}
