package expanded

import (
	ui "github.com/gizak/termui"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

var paramInfo = []string{"name", "description", "url", "path", "type", "contents"}

type TaskParameter struct {
	*ui.Table
	data map[string]string
}

func NewTaskParameters(t []*tes.TaskParameter, label string) *TaskParameter {
	p := ui.NewTable()
	p.Height = 4
	p.Width = colWidth[0]
	p.FgColor = ui.ThemeAttr("par.text.fg")
	p.Separator = false
	p.BorderLabel = label

	i := &TaskParameter{p, make(map[string]string)}
	// i.Set("name", t.Name)
	// i.Set("description", t.Description)
	// i.Set("url", t.Url)
	// i.Set("path", t.Path)
	// i.Set("path", t.Type)

	return i
}

func (w *TaskParameter) Set(k, v string) {
	w.data[k] = v
	// rebuild rows
	w.Rows = [][]string{}
	for _, k := range paramInfo {
		if v, ok := w.data[k]; ok {
			w.Rows = append(w.Rows, []string{k, v})
		}
	}
	w.Height = len(w.Rows) + 2
}
