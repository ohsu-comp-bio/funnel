package expanded

import (
	ui "github.com/gizak/termui"
	"strings"
)

type JSON struct {
	*ui.Par
}

func NewJSON(s string) *JSON {
	p := ui.NewPar(s)
	p.Border = true
	p.BorderLabel = "TASK"
	p.Height = strings.Count(s, "\n") + 3
	p.Width = ui.TermWidth()
	return &JSON{p}
}

func (j *JSON) Set(s string) {
	j.Text = s
	j.Height = strings.Count(s, "\n") + 3
}
