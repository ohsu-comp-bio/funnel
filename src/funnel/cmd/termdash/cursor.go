// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package termdash

import (
	ui "github.com/gizak/termui"
	"math"
)

type GridCursor struct {
	selectedID string
	filtered   TaskWidgets
	tSource    TesTaskSource
}

func NewGridCursor(tesHTTPServerAddress string) *GridCursor {
	return &GridCursor{
		tSource: NewTaskSource(tesHTTPServerAddress),
	}
}

func (gc *GridCursor) Len() int { return len(gc.filtered) }

func (gc *GridCursor) Selected() *TaskWidget {
	idx := gc.Idx()
	if idx < gc.Len() {
		return gc.filtered[idx]
	}
	return nil
}

// Refresh a single task
func (gc *GridCursor) RefreshTask(id string) *TaskWidget {
	return gc.tSource.Get(id)
}

// Refresh task list
func (gc *GridCursor) RefreshTaskList() (lenChanged bool) {
	oldLen := gc.Len()

	// Tasks filtered by display bool
	gc.filtered = TaskWidgets{}
	var cursorVisible bool
	for _, t := range gc.tSource.All() {
		if t.display {
			if t.Task.Id == gc.selectedID {
				t.Widgets.ID.Highlight()
				cursorVisible = true
			}
			gc.filtered = append(gc.filtered, t)
		}
	}

	if oldLen != gc.Len() {
		lenChanged = true
	}

	if !cursorVisible {
		gc.Reset()
	}

	if gc.selectedID == "" {
		gc.Reset()
	}

	return lenChanged
}

// Set an initial cursor position, if possible
func (gc *GridCursor) Reset() {
	for _, t := range gc.tSource.All() {
		t.Widgets.ID.UnHighlight()
	}
	if gc.Len() > 0 {
		gc.selectedID = gc.filtered[0].Task.Id
		gc.filtered[0].Widgets.ID.Highlight()
	}
}

// Return current cursor index
func (gc *GridCursor) Idx() int {
	for n, t := range gc.filtered {
		if t.Task.Id == gc.selectedID {
			return n
		}
	}
	gc.Reset()
	return 0
}

func (gc *GridCursor) ScrollPage() {
	// skip scroll if no need to page
	if gc.Len() < cGrid.MaxRows() {
		cGrid.Offset = 0
		return
	}

	idx := gc.Idx()

	// page down
	if idx >= cGrid.Offset+cGrid.MaxRows() {
		cGrid.Offset++
		cGrid.Align()
	}
	// page up
	if idx < cGrid.Offset {
		cGrid.Offset--
		cGrid.Align()
	}

}

func (gc *GridCursor) Up() {
	idx := gc.Idx()
	if idx <= 0 { // already at top
		return
	}
	active := gc.filtered[idx]
	next := gc.filtered[idx-1]

	active.Widgets.ID.UnHighlight()
	gc.selectedID = next.Task.Id
	next.Widgets.ID.Highlight()

	gc.ScrollPage()
	ui.Render(cGrid)
}

func (gc *GridCursor) Down() {
	idx := gc.Idx()
	if idx >= gc.Len()-1 { // already at bottom
		return
	}
	active := gc.filtered[idx]
	next := gc.filtered[idx+1]

	active.Widgets.ID.UnHighlight()
	gc.selectedID = next.Task.Id
	next.Widgets.ID.Highlight()

	gc.ScrollPage()
	ui.Render(cGrid)
}

func (gc *GridCursor) PgUp() {
	idx := gc.Idx()
	if idx <= 0 { // already at top
		return
	}

	var nextidx int
	nextidx = int(math.Max(0.0, float64(idx-cGrid.MaxRows())))
	cGrid.Offset = int(math.Max(float64(cGrid.Offset-cGrid.MaxRows()),
		float64(0)))

	active := gc.filtered[idx]
	next := gc.filtered[nextidx]

	active.Widgets.ID.UnHighlight()
	gc.selectedID = next.Task.Id
	next.Widgets.ID.Highlight()

	cGrid.Align()
	ui.Render(cGrid)
}

func (gc *GridCursor) PgDown() {
	idx := gc.Idx()
	if idx >= gc.Len()-1 { // already at bottom
		return
	}

	var nextidx int
	nextidx = int(math.Min(float64(gc.Len()-1),
		float64(idx+cGrid.MaxRows())))
	cGrid.Offset = int(math.Min(float64(cGrid.Offset+cGrid.MaxRows()),
		float64(gc.Len()-cGrid.MaxRows())))

	active := gc.filtered[idx]
	next := gc.filtered[nextidx]

	active.Widgets.ID.UnHighlight()
	gc.selectedID = next.Task.Id
	next.Widgets.ID.Highlight()

	cGrid.Align()
	ui.Render(cGrid)
}
