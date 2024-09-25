// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package config

var (
	GlobalParams   []*Param
	GlobalSwitches []*Switch
)

func Init() {
	GlobalParams = append(GlobalParams, params...)
	GlobalSwitches = append(GlobalSwitches, switches...)
}
