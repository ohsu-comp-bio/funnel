// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package compact

import (
	ui "github.com/gizak/termui"
)

type Header struct {
	X, Y   int
	Width  int
	Height int
	pars   []*ui.Par
}

func NewHeader() *Header {
	fields := []string{"", "ID", "STATE", "NAME", "DESCRIPTION"}
	ch := &Header{}
	ch.Height = 2
	for _, f := range fields {
		ch.addFieldPar(f)
	}
	return ch
}

func (ch *Header) GetHeight() int {
	return ch.Height
}

func (ch *Header) SetWidth(w int) {
	if w == ch.Width {
		return
	}
	x := ch.X
	autoWidth := calcWidth(w)
	for n, col := range ch.pars {
		// set column to static width
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
	ch.Width = w
}

func (ch *Header) SetX(x int) {
	ch.X = x
}

func (ch *Header) SetY(y int) {
	for _, p := range ch.pars {
		p.SetY(y)
	}
	ch.Y = y
}

func (ch *Header) Buffer() ui.Buffer {
	buf := ui.NewBuffer()
	for _, p := range ch.pars {
		buf.Merge(p.Buffer())
	}
	return buf
}

func (ch *Header) addFieldPar(s string) {
	p := ui.NewPar(s)
	p.Height = ch.Height
	p.Border = false
	ch.pars = append(ch.pars, p)
}
