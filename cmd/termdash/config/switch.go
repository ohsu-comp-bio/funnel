// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package config

// defaults
var switches = []*Switch{
	{
		Key:   "sortReversed",
		Val:   true,
		Label: "Reverse Sort Order",
	},
	{
		Key:   "allTasks",
		Val:   false,
		Label: "Clear All Filters",
	},
	{
		Key:   "enableHeader",
		Val:   true,
		Label: "Enable Status Header",
	},
}

type Switch struct {
	Key   string
	Val   bool
	Label string
}

// Return Switch by key
func GetSwitch(k string) *Switch {
	for _, sw := range GlobalSwitches {
		if sw.Key == k {
			return sw
		}
	}
	return &Switch{} // default
}

// Return Switch value by key
func GetSwitchVal(k string) bool {
	return GetSwitch(k).Val
}

// Toggle a boolean switch
func Toggle(k string) {
	sw := GetSwitch(k)
	newVal := !sw.Val
	sw.Val = newVal
}
