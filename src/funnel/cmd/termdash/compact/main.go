// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package compact

import (
	"funnel/proto/tes"
	ui "github.com/gizak/termui"
)

//var log = logging.Init()

type Compact struct {
	Status *Status
	ID     *TextCol
	State  *TextCol
	Name   *TextCol
	Desc   *TextCol
	X, Y   int
	Width  int
	Height int
}

func NewCompact(t *tes.Task) *Compact {
	// truncate task id
	id := t.Id
	if len(id) > 12 {
		id = id[:12]
	}
	row := &Compact{
		Status: NewStatus(),
		ID:     NewTextCol(id),
		State:  NewTextCol(t.State.String()),
		Name:   NewTextCol(t.Name),
		Desc:   NewTextCol(t.Description),
		X:      0,
		Height: 1,
	}
	row.Status.Set(t.State.String())
	return row
}

func (row *Compact) GetHeight() int {
	return row.Height
}

func (row *Compact) SetX(x int) {
	row.X = x
}

func (row *Compact) SetY(y int) {
	if y == row.Y {
		return
	}
	for _, col := range row.all() {
		col.SetY(y)
	}
	row.Y = y
}

func (row *Compact) SetWidth(width int) {
	if width == row.Width {
		return
	}
	x := row.X
	autoWidth := calcWidth(width)
	for n, col := range row.all() {
		if colWidths[n] != 0 {
			col.SetX(x)
			col.SetWidth(colWidths[n])
			x += colWidths[n]
			continue
		}
		col.SetX(x)
		col.SetWidth(autoWidth)
		x += autoWidth + colSpacing
	}
	row.Width = width
}

func (row *Compact) Buffer() ui.Buffer {
	buf := ui.NewBuffer()
	buf.Merge(row.Status.Buffer())
	buf.Merge(row.ID.Buffer())
	buf.Merge(row.State.Buffer())
	buf.Merge(row.Name.Buffer())
	buf.Merge(row.Desc.Buffer())
	return buf
}

func (row *Compact) all() []ui.GridBufferer {
	return []ui.GridBufferer{
		row.Status,
		row.ID,
		row.State,
		row.Name,
		row.Desc,
	}
}
