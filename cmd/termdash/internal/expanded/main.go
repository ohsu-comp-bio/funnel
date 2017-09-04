// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package expanded

import (
	ui "github.com/gizak/termui"
	"github.com/golang/protobuf/jsonpb"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"strings"
)

var (
	sizeError = termSizeError()
	colWidth  = [2]int{65, 0} // left,right column width
	marshaler = &jsonpb.Marshaler{
		EnumsAsInts:  false,
		EmitDefaults: false,
		Indent:       "  ",
	}
)

type Expanded struct {
	TaskJSON *JSON
	// TaskInfo    *TaskInfo
	// TaskInputs  *TaskParameter
	// TaskOutputs *TaskParameter
	// TaskResources *TaskResources
	// TaskExecutors *TaskExecutors
	// TaskLogs      *TaskLogs
	X, Y  int
	Width int
}

func NewExpanded(t *tes.Task) *Expanded {
	ts, _ := marshaler.MarshalToString(t)
	ts = strings.Replace(ts, `\n`, "\n", -1)
	return &Expanded{
		TaskJSON: NewJSON(ts),
		// TaskInfo:    NewTaskInfo(t),
		// TaskInputs:  NewTaskParameters(t.Inputs, "INPUTS"),
		// TaskOutputs: NewTaskParameters(t.Outputs, "OUTPUTS"),
		// TaskExecutors: NewTaskExecutors(t.Executors),
		// TaskResources: NewTaskExecutors(t.Resources),
		// TaskLogs: NewTaskLogs(t.Logs),
		Width: ui.TermWidth(),
	}
}

func (e *Expanded) Update(t *tes.Task) {
	ts, _ := marshaler.MarshalToString(t)
	ts = strings.Replace(ts, `\n`, "\n", -1)
	e.TaskJSON.Set(ts)
	// e.TaskInfo.Set(t)
	// e.TaskInputs.Set(t.Inputs)
	// e.TaskOutputs.Set(t.Outputs)
	// e.TaskExecutors.Set(t.Executors)
	// e.TaskResources .Set(t.Resources)
	// e.TaskLogs .Set(t.Logs)
	e.Width = ui.TermWidth()
}

func (e *Expanded) Up() {
	if e.Y < 0 {
		e.Y++
		e.Align()
		ui.Render(e)
	}
}

func (e *Expanded) Down() {
	if e.Y > (ui.TermHeight() - e.GetHeight()) {
		e.Y--
		e.Align()
		ui.Render(e)
	}
}

func (e *Expanded) SetWidth(w int) {
	e.Width = w
}

// Return total column height
func (e *Expanded) GetHeight() (h int) {
	h += e.TaskJSON.Height
	// h += e.TaskInfo.Height
	// h += e.TaskInputs.Height
	// h += e.TaskOutputs.Height
	// h += e.TaskExecutors.Height
	// h += e.TaskResources.Height
	// h += e.TaskLogs.Height
	return h
}

func (e *Expanded) Align() {
	// reset offset if needed
	if e.GetHeight() <= ui.TermHeight() {
		e.Y = 0
	}

	y := e.Y
	for _, i := range e.all() {
		i.SetY(y)
		y += i.GetHeight()
	}

	if e.Width > colWidth[0] {
		colWidth[1] = e.Width - (colWidth[0] + 1)
	}
}

func (e *Expanded) Buffer() ui.Buffer {
	buf := ui.NewBuffer()
	if e.Width < (colWidth[0] + colWidth[1]) {
		ui.Clear()
		buf.Merge(sizeError.Buffer())
		return buf
	}
	buf.Merge(e.TaskJSON.Buffer())
	// buf.Merge(e.TaskInfo.Buffer())
	// buf.Merge(e.TaskInputs.Buffer())
	// buf.Merge(e.TaskOutputs.Buffer())
	// buf.Merge(e.TaskExecutors.Buffer())
	// buf.Merge(e.TaskResources.Buffer())
	// buf.Merge(e.TaskLogs.Buffer())
	return buf
}

func (e *Expanded) all() []ui.GridBufferer {
	return []ui.GridBufferer{
		e.TaskJSON,
		// e.TaskInfo,
		// e.TaskInputs,
		// e.TaskOutputs,
		// e.TaskExecutors,
		// e.TaskResources,
		// e.TaskLogs,
	}
}

func termSizeError() *ui.Par {
	p := ui.NewPar("screen too small!")
	p.Height = 1
	p.Width = 20
	p.Border = false
	return p
}
