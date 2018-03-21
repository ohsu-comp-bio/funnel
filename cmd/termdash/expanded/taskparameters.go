package expanded

import (
	ui "github.com/gizak/termui"
	"github.com/ohsu-comp-bio/funnel/tes"
)

var inputInfo = []string{"name", "description", "url", "path", "type", "content"}

type TaskInput struct {
	*ui.Table
	data map[string]string
}

func NewTaskInputs(t []*tes.Input, label string) *TaskInput {
	p := ui.NewTable()
	p.Height = 4
	p.Width = colWidth[0]
	p.FgColor = ui.ThemeAttr("par.text.fg")
	p.Separator = false
	p.BorderLabel = label

	i := &TaskInput{p, make(map[string]string)}
	// i.Set("name", t.Name)
	// i.Set("description", t.Description)
	// i.Set("url", t.Url)
	// i.Set("path", t.Path)
	// i.Set("path", t.Type)

	return i
}

func (w *TaskInput) Set(k, v string) {
	w.data[k] = v
	// rebuild rows
	w.Rows = [][]string{}
	for _, k := range inputInfo {
		if v, ok := w.data[k]; ok {
			w.Rows = append(w.Rows, []string{k, v})
		}
	}
	w.Height = len(w.Rows) + 2
}

var outputInfo = []string{"name", "description", "url", "path", "type", "content"}

type TaskOutput struct {
	*ui.Table
	data map[string]string
}

func NewTaskOutputs(t []*tes.Output, label string) *TaskOutput {
	p := ui.NewTable()
	p.Height = 4
	p.Width = colWidth[0]
	p.FgColor = ui.ThemeAttr("par.text.fg")
	p.Separator = false
	p.BorderLabel = label

	i := &TaskOutput{p, make(map[string]string)}
	// i.Set("name", t.Name)
	// i.Set("description", t.Description)
	// i.Set("url", t.Url)
	// i.Set("path", t.Path)
	// i.Set("path", t.Type)

	return i
}

func (w *TaskOutput) Set(k, v string) {
	w.data[k] = v
	// rebuild rows
	w.Rows = [][]string{}
	for _, k := range outputInfo {
		if v, ok := w.data[k]; ok {
			w.Rows = append(w.Rows, []string{k, v})
		}
	}
	w.Height = len(w.Rows) + 2
}
