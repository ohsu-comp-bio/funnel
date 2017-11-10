package expanded

import (
	"fmt"
	ui "github.com/gizak/termui"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"strings"
)

var displayInfo = []string{"id", "state", "name", "description", "tags"}

type TaskInfo struct {
	*ui.Table
	data map[string]string
}

func NewTaskInfo(t *tes.Task) *TaskInfo {
	p := ui.NewTable()
	p.Height = 5
	p.Width = colWidth[0]
	p.FgColor = ui.ThemeAttr("par.text.fg")
	p.Separator = false

	i := &TaskInfo{p, make(map[string]string)}
	i.Set("id", t.Id)
	i.Set("state", t.State.String())
	i.Set("name", t.Name)
	i.Set("description", t.Description)
	var tags []string
	for k, v := range t.Tags {
		tags = append(tags, fmt.Sprintf("%s: %s", k, v))
	}
	i.Set("tags", strings.Join(tags, ", "))
	return i
}

func (w *TaskInfo) Set(k, v string) {
	w.data[k] = v
	// rebuild rows
	w.Rows = [][]string{}
	for _, k := range displayInfo {
		if v, ok := w.data[k]; ok {
			w.Rows = append(w.Rows, []string{k, v})
		}
	}
	w.Height = len(w.Rows) + 2
}
