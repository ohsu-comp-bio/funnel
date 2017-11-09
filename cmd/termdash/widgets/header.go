// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package widgets

import (
	"fmt"
	"time"

	ui "github.com/gizak/termui"
)

type TermDashHeader struct {
	Error    *ui.List
	Time     *ui.Par
	Previous *ui.Par
	Next     *ui.Par
	Filter   *ui.Par
	help     *ui.Par
	bg       *ui.Par
}

func NewTermDashHeader() *TermDashHeader {
	return &TermDashHeader{
		Error:    headerError(""),
		Time:     headerPar(0, 0, timeStr()),
		Previous: headerPar(ui.TermWidth()/2-12, 0, ""),
		Next:     headerPar(ui.TermWidth()/2+1, 0, ""),
		Filter:   headerPar(0, 1, ""),
		help:     headerPar(ui.TermWidth()-20, 0, "press [h] for help"),
		bg:       headerBg(),
	}
}

func (c *TermDashHeader) Buffer() ui.Buffer {
	buf := ui.NewBuffer()
	buf.Merge(c.bg.Buffer())
	c.Time.Text = timeStr()
	buf.Merge(c.Time.Buffer())
	buf.Merge(c.Previous.Buffer())
	buf.Merge(c.Next.Buffer())
	if c.Filter.Text != "" {
		buf.Merge(c.Filter.Buffer())
	}
	buf.Merge(c.help.Buffer())
	if len(c.Error.Items) > 0 {
		buf.Merge(c.Error.Buffer())
	}
	return buf
}

func (c *TermDashHeader) Align() {
	c.bg.SetWidth(ui.TermWidth())
}

func (c *TermDashHeader) Height() int {
	return c.bg.Height
}

func headerBg() *ui.Par {
	bg := ui.NewPar("")
	bg.X = 0
	bg.Height = 1
	bg.Width = ui.TermWidth()
	bg.Border = false
	bg.Bg = ui.ThemeAttr("header.bg")
	return bg
}

func (c *TermDashHeader) SetPrevious(set bool) {
	if set {
		c.Previous.Text = "<- Previous"
	} else {
		c.Previous.Text = ""
	}
}

func (c *TermDashHeader) SetNext(set bool) {
	if set {
		c.Next.Text = "Next ->"
	} else {
		c.Next.Text = ""
	}
}

func (c *TermDashHeader) SetFilter(val string) {
	switch {
	case val == "" && len(c.Error.Items) == 0:
		c.bg.Height = 1
		c.Filter.Text = ""
	case val == "" && len(c.Error.Items) > 0:
		c.Filter.Text = ""
	case val != "" && len(c.Error.Items) == 0:
		c.bg.Height = 2
		val = fmt.Sprintf("filter: %s", val)
		c.Filter.X = ui.TermWidth()/2 - len(val)/2
		c.Filter.Text = val
	case val != "" && len(c.Error.Items) > 0:
		c.bg.Height = 4
		val = fmt.Sprintf("filter: %s", val)
		c.Filter.X = ui.TermWidth()/2 - len(val)/2
		c.Filter.Text = val
	}
}

func (c *TermDashHeader) SetError(val string) {
	switch {
	case val == "" && c.Filter.Text == "":
		c.bg.Height = 1
		c.Error.Items = []string{}
	case val == "" && c.Filter.Text != "":
		c.Error.Items = []string{}
	case val != "" && c.Filter.Text == "":
		c.Error.Y = 1
		c.bg.Height = 3
		c.Error.Items = []string{fmt.Sprintf("ERROR: %s", val)}
	case val != "" && c.Filter.Text != "":
		c.Error.Y = 2
		c.bg.Height = 4
		c.Error.Items = []string{fmt.Sprintf("ERROR: %s", val)}
	}
}

func timeStr() string {
	ts := time.Now().Local().Format("15:04:05 MST")
	return fmt.Sprintf("funnel - %s", ts)
}

func headerPar(x, y int, s string) *ui.Par {
	p := ui.NewPar(fmt.Sprintf(" %s", s))
	p.X = x
	p.Y = y
	p.Border = false
	p.Height = 1
	p.Width = 50
	p.Bg = ui.ThemeAttr("header.bg")
	p.TextFgColor = ui.ThemeAttr("header.fg")
	p.TextBgColor = ui.ThemeAttr("header.bg")
	return p
}

func headerError(s string) *ui.List {
	p := ui.NewList()
	if s != "" {
		p.Items = []string{s}
	}
	p.X = 0
	p.Y = 1
	p.Border = false
	p.Height = 2
	p.Width = ui.TermWidth() - 1
	p.Overflow = "wrap"
	p.Bg = ui.ThemeAttr("header.bg")
	p.ItemFgColor = ui.ColorRed
	p.ItemBgColor = ui.ThemeAttr("header.bg")
	return p
}
