package config

// defaults
var params = []*Param{
	{
		Key:   "filterStr",
		Val:   "",
		Label: "Task Name or ID Filter",
	},
	{
		Key:   "sortField",
		Val:   "id",
		Label: "Task Sort Field",
	},
}

type Param struct {
	Key   string
	Val   string
	Label string
}

// Get Param by key
func Get(k string) *Param {
	for _, p := range GlobalParams {
		if p.Key == k {
			return p
		}
	}
	return &Param{}
}

// Get Param value by key
func GetVal(k string) string {
	return Get(k).Val
}

// Set param value
func Update(k, v string) {
	p := Get(k)
	p.Val = v
}
