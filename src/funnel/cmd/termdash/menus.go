// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package termdash

import (
	"funnel/cmd/termdash/config"
	"funnel/cmd/termdash/widgets"
	"funnel/cmd/termdash/widgets/menu"
	ui "github.com/gizak/termui"
)

var helpDialog = []menu.Item{
	{Label: "[a] - toggle display of all tasks", Val: ""},
	{Label: "[f] - filter displayed tasks", Val: ""},
	{Label: "[h] - open this help dialog", Val: ""},
	{Label: "[H] - toggle dashboard header", Val: ""},
	{Label: "[s] - select task sort field", Val: ""},
	{Label: "[r] - reverse task sort order", Val: ""},
	{Label: "[q] - exit funnel dashboard", Val: ""},
}

func HelpMenu() {
	ui.Clear()
	ui.DefaultEvtStream.ResetHandlers()
	defer ui.DefaultEvtStream.ResetHandlers()

	m := menu.NewMenu()
	m.BorderLabel = "Help"
	m.AddItems(helpDialog...)
	ui.Render(m)
	ui.Handle("/sys/kbd/", func(ui.Event) {
		ui.StopLoop()
	})
	ui.Loop()
}

func FilterMenu() {
	ui.DefaultEvtStream.ResetHandlers()
	defer ui.DefaultEvtStream.ResetHandlers()

	i := widgets.NewInput()
	i.BorderLabel = "Filter"
	i.SetY(ui.TermHeight() - i.Height)
	i.Data = config.GetVal("filterStr")
	ui.Render(i)

	// refresh container rows on input
	stream := i.Stream()
	go func() {
		for s := range stream {
			config.Update("filterStr", s)
			RefreshDisplay()
			ui.Render(i)
		}
	}()

	i.InputHandlers()
	ui.Handle("/sys/kbd/<escape>", func(ui.Event) {
		config.Update("filterStr", "")
		ui.StopLoop()
	})
	ui.Handle("/sys/kbd/<enter>", func(ui.Event) {
		config.Update("filterStr", i.Data)
		ui.StopLoop()
	})
	ui.Loop()
}

func SortMenu() {
	ui.Clear()
	ui.DefaultEvtStream.ResetHandlers()
	defer ui.DefaultEvtStream.ResetHandlers()

	m := menu.NewMenu()
	m.Selectable = true
	m.SortItems = true
	m.BorderLabel = "Sort Field"

	for _, field := range SortFields() {
		m.AddItems(menu.Item{Label: field, Val: ""})
	}

	// set cursor position to current sort field
	m.SetCursor(config.GetVal("sortField"))

	HandleKeys("up", m.Up)
	HandleKeys("down", m.Down)
	HandleKeys("exit", ui.StopLoop)

	ui.Handle("/sys/kbd/<enter>", func(ui.Event) {
		config.Update("sortField", m.SelectedItem().Val)
		ui.StopLoop()
	})

	ui.Render(m)
	ui.Loop()
}
