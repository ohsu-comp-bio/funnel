// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package compact

import (
	"fmt"

	ui "github.com/gizak/termui"
)

const (
	mark        = string('\u25C9')
	vBar        = string('\u25AE')
	statusWidth = 3
)

// Status indicator
type Status struct {
	*ui.Par
}

func NewStatus() *Status {
	p := ui.NewPar(mark)
	p.Border = false
	p.Height = 1
	p.Width = statusWidth
	return &Status{p}
}

func (s *Status) Set(val string) {
	// defaults
	text := mark
	color := ui.ColorDefault

	switch val {
	case "QUEUED":
		color = ui.ColorWhite
	case "RUNNING", "INITIALIZING":
		color = ui.ColorGreen
	case "COMPLETE", "ERROR", "CANCELED", "SYSTEM_ERRROR":
		color = ui.ColorRed
	case "UNKNOWN", "PAUSED":
		text = fmt.Sprintf("%s%s", vBar, vBar)
	}

	s.Text = text
	s.TextFgColor = color
}
