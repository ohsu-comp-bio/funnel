// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package compact

// Common helper functions

const colSpacing = 1

// per-column width. 0 == auto width
var colWidths = []int{
	3, // status
	0, // id
	0, // state
	0, // name
	0, // desc
}

// Calculate per-column width, given total width
func calcWidth(width int) int {
	spacing := colSpacing * len(colWidths)
	var staticCols int
	for _, w := range colWidths {
		width -= w
		if w == 0 {
			staticCols++
		}
	}
	return (width - spacing) / staticCols
}
