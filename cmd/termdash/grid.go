// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package termdash

import (
	ui "github.com/gizak/termui"
	"github.com/ohsu-comp-bio/funnel/cmd/termdash/config"
	"github.com/ohsu-comp-bio/funnel/cmd/termdash/expanded"
)

func RedrawRows(clr bool) {
	// reinit body rows
	cGrid.Clear()

	// build layout
	y := 1
	if config.GetSwitchVal("enableHeader") {
		header.SetPrevious(cursor.PreviousPageExists())
		header.SetNext(cursor.NextPageExists())
		header.SetFilter(config.GetVal("filterStr"))
		y += header.Height()
	}
	cGrid.SetY(y)

	for _, c := range cursor.filtered {
		cGrid.AddRows(c.Widgets)
	}

	if clr {
		ui.Clear()
	}
	if config.GetSwitchVal("enableHeader") {
		ui.Render(header)
	}
	cGrid.Align()
	ui.Render(cGrid)
}

func ExpandView(t *TaskWidget) {
	ui.Clear()
	ui.DefaultEvtStream.ResetHandlers()
	defer ui.DefaultEvtStream.ResetHandlers()

	ex := expanded.NewExpanded(t.Task)
	ex.Align()
	ui.Render(ex)

	HandleKeys("up", ex.Up)
	HandleKeys("down", ex.Down)
	HandleKeys("exit", ui.StopLoop)
	ui.Handle("/timer/1s", func(e ui.Event) {
		task, err := cursor.RefreshTask(t.Task.Id)
		if err != nil {
			ex.DisplayError(err)
		} else {
			ex.Update(task.Task)
		}
		ui.Clear()
		ex.Align()
		ui.Render(ex)
	})
	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		ex.SetWidth(ui.TermWidth())
		ex.Align()
	})

	ui.Loop()
}

func RefreshDisplay() {
	needsClear := cursor.RefreshTaskList(false, false)
	RedrawRows(needsClear)
}

func NextPage() {
	cursor.RefreshTaskList(false, true)
	RedrawRows(true)
}

func PreviousPage() {
	cursor.RefreshTaskList(true, false)
	RedrawRows(true)
}

func Display() bool {
	var menu func()
	var expand bool

	cGrid.SetWidth(ui.TermWidth())

	// initial draw
	header.Align()
	cursor.RefreshTaskList(false, false)
	RedrawRows(true)

	HandleKeys("up", cursor.Up)
	HandleKeys("down", cursor.Down)
	HandleKeys("left", PreviousPage)
	HandleKeys("right", NextPage)
	HandleKeys("exit", ui.StopLoop)
	HandleKeys("help", func() {
		menu = HelpMenu
		ui.StopLoop()
	})

	ui.Handle("/sys/kbd/<enter>", func(ui.Event) {
		expand = true
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/a", func(ui.Event) {
		config.Toggle("allTasks")
		RefreshDisplay()
	})
	ui.Handle("/sys/kbd/f", func(ui.Event) {
		menu = FilterMenu
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/H", func(ui.Event) {
		config.Toggle("enableHeader")
		RedrawRows(true)
	})
	ui.Handle("/sys/kbd/r", func(e ui.Event) {
		config.Toggle("sortReversed")
	})
	ui.Handle("/sys/kbd/s", func(ui.Event) {
		menu = SortMenu
		ui.StopLoop()
	})
	ui.Handle("/timer/1s", func(e ui.Event) {
		RefreshDisplay()
	})
	ui.Handle("/sys/wnd/resize", func(e ui.Event) {
		header.Align()
		cursor.ScrollPage()
		cGrid.SetWidth(ui.TermWidth())
		RedrawRows(true)
	})
	ui.Loop()

	if menu != nil {
		menu()
		return false
	}

	if expand {
		task := cursor.Selected()
		if task != nil {
			ExpandView(task)
		}
		return false
	}

	return true
}
