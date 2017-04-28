// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package widgets

import (
	"fmt"
	"time"

	ui "github.com/gizak/termui"
)

type TermDashHeader struct {
	Time   *ui.Par
	Count  *ui.Par
	Filter *ui.Par
	help   *ui.Par
	bg     *ui.Par
}

func NewTermDashHeader() *TermDashHeader {
	return &TermDashHeader{
		Time:   headerPar(0, timeStr()),
		Count:  headerPar(33, "-"),
		Filter: headerPar(50, ""),
		help:   headerPar(ui.TermWidth()-19, "press [h] for help"),
		bg:     headerBg(),
	}
}

func (c *TermDashHeader) Buffer() ui.Buffer {
	buf := ui.NewBuffer()
	c.Time.Text = timeStr()
	buf.Merge(c.bg.Buffer())
	buf.Merge(c.Time.Buffer())
	buf.Merge(c.Count.Buffer())
	buf.Merge(c.Filter.Buffer())
	buf.Merge(c.help.Buffer())
	return buf
}

func (c *TermDashHeader) Align() {
	c.bg.SetWidth(ui.TermWidth() - 1)
}

func (c *TermDashHeader) Height() int {
	return c.bg.Height
}

func headerBgBordered() *ui.Par {
	bg := ui.NewPar("")
	bg.X = 1
	bg.Height = 3
	bg.Bg = ui.ThemeAttr("header.bg")
	return bg
}

func headerBg() *ui.Par {
	bg := ui.NewPar("")
	bg.X = 1
	bg.Height = 1
	bg.Border = false
	bg.Bg = ui.ThemeAttr("header.bg")
	return bg
}

func (c *TermDashHeader) SetCount(val int) {
	c.Count.Text = fmt.Sprintf("%d tasks", val)
}

func (c *TermDashHeader) SetFilter(val string) {
	if val == "" {
		c.Filter.Text = ""
	} else {
		c.Filter.Text = fmt.Sprintf("filter: %s", val)
	}
}

func timeStr() string {
	ts := time.Now().Local().Format("15:04:05 MST")
	return fmt.Sprintf("funnel - %s", ts)
}

func headerPar(x int, s string) *ui.Par {
	p := ui.NewPar(fmt.Sprintf(" %s", s))
	p.X = x
	p.Border = false
	p.Height = 1
	p.Width = 24
	p.Bg = ui.ThemeAttr("header.bg")
	p.TextFgColor = ui.ThemeAttr("header.fg")
	p.TextBgColor = ui.ThemeAttr("header.bg")
	return p
}
