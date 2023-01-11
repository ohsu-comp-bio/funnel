// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package compact

import (
	ui "github.com/gizak/termui"
)

var header *Header

type Grid struct {
	ui.GridBufferer
	Rows   []ui.GridBufferer
	X, Y   int
	Width  int
	Height int
	Offset int // starting row offset
}

func NewGrid() *Grid {
	header = NewHeader() // init column header
	return &Grid{}
}

func (cg *Grid) Align() {
	y := cg.Y
	if cg.Offset >= len(cg.Rows) {
		cg.Offset = 0
	}
	if cg.Offset < 0 {
		cg.Offset = 0
	}
	// update row ypos, width recursively
	for _, r := range cg.pageRows() {
		r.SetY(y)
		y += r.GetHeight()
		r.SetWidth(cg.Width)
	}
}

func (cg *Grid) Clear()         { cg.Rows = []ui.GridBufferer{} }
func (cg *Grid) GetHeight() int { return len(cg.Rows) + header.Height }
func (cg *Grid) SetX(x int)     { cg.X = x }
func (cg *Grid) SetY(y int)     { cg.Y = y }
func (cg *Grid) SetWidth(w int) { cg.Width = w }
func (cg *Grid) MaxRows() int   { return ui.TermHeight() - header.Height - cg.Y }

func (cg *Grid) pageRows() (rows []ui.GridBufferer) {
	rows = append(rows, header)
	rows = append(rows, cg.Rows[cg.Offset:]...)
	return rows
}

func (cg *Grid) Buffer() ui.Buffer {
	buf := ui.NewBuffer()
	for _, r := range cg.pageRows() {
		buf.Merge(r.Buffer())
	}
	return buf
}

func (cg *Grid) AddRows(rows ...ui.GridBufferer) {
	cg.Rows = append(cg.Rows, rows...)
}
