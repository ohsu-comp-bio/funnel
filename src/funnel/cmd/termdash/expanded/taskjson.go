package expanded

import (
	ui "github.com/gizak/termui"
	"strings"
)

type Json struct {
	*ui.Par
}

func NewJson(s string) *Json {
	p := ui.NewPar(s)
	p.Border = true
	p.BorderLabel = "TASK"
	p.Height = strings.Count(s, "\n") + 3
	p.Width = ui.TermWidth()
	return &Json{p}
}

func (j *Json) Set(s string) {
	j.Text = s
	j.Height = strings.Count(s, "\n") + 3
}
