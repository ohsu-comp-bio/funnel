package expanded

import (
	"strings"

	ui "github.com/gizak/termui"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

type JSON struct {
	*ui.Par
}

func NewJSON(t *tes.Task) *JSON {
	ts, _ := tes.MarshalToString(t)
	ts = strings.Replace(ts, `\n`, "\n", -1)

	p := ui.NewPar(ts)
	p.TextFgColor = ui.ColorWhite
	p.Border = true
	p.BorderLabel = "TASK"
	p.Height = strings.Count(ts, "\n") + 3
	p.Width = ui.TermWidth()
	return &JSON{p}
}

func (j *JSON) Set(t *tes.Task) {
	ts, _ := tes.MarshalToString(t)
	ts = strings.Replace(ts, `\n`, "\n", -1)

	j.TextFgColor = ui.ColorWhite
	j.Text = ts
	j.Height = strings.Count(ts, "\n") + 3
}

func (j *JSON) SetErrorMessage(err error) {
	j.TextFgColor = ui.ColorRed
	j.Text = err.Error()
	j.Height = 10
}
