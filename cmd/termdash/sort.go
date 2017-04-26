// Copied and modified from: https://github.com/bcicen/ctop
// MIT License - Copyright (c) 2017 VektorLab

package termdash

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/cmd/termdash/config"
	"regexp"
)

type sortMethod func(c1, c2 *TaskWidget) bool

var idSorter = func(c1, c2 *TaskWidget) bool {
	return c1.Task.Id < c2.Task.Id
}

var stateSorter = func(c1, c2 *TaskWidget) bool {
	if c1.Task.State.String() == c2.Task.State.String() {
		return nameSorter(c1, c2)
	}
	return c1.Task.State.String() < c2.Task.State.String()
}

var nameSorter = func(c1, c2 *TaskWidget) bool {
	return c1.Task.Name < c2.Task.Name
}

var descSorter = func(c1, c2 *TaskWidget) bool {
	return c1.Task.Description < c2.Task.Description
}

var Sorters = map[string]sortMethod{
	"id":          idSorter,
	"state":       stateSorter,
	"name":        nameSorter,
	"description": descSorter,
}

func SortFields() (fields []string) {
	for k := range Sorters {
		fields = append(fields, k)
	}
	return fields
}

type TaskWidgets []*TaskWidget

func (a TaskWidgets) Len() int      { return len(a) }
func (a TaskWidgets) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a TaskWidgets) Less(i, j int) bool {
	f := Sorters[config.GetVal("sortField")]
	if config.GetSwitchVal("sortReversed") {
		return f(a[j], a[i])
	}
	return f(a[i], a[j])
}

func (a TaskWidgets) Filter() {
	filter := config.GetVal("filterStr")
	re := regexp.MustCompile(fmt.Sprintf(".*%s", filter))

	for _, t := range a {
		// Apply filter
		fi := re.FindAllString(t.Task.Id, 1) == nil
		fn := re.FindAllString(t.Task.Name, 1) == nil
		fs := re.FindAllString(t.Task.State.String(), 1) == nil
		fd := re.FindAllString(t.Task.Description, 1) == nil
		if fi && fn && fs && fd {
			t.display = false
		}

		if config.GetSwitchVal("allTasks") {
			t.display = true
		}
	}
}
