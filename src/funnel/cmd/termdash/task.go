package termdash

import (
	"funnel/cmd/termdash/compact"
	"funnel/proto/tes"
)

type TaskWidget struct {
	Task    *tes.Task
	Widgets *compact.Compact
	display bool
}

func NewTaskWidget(t *tes.Task) *TaskWidget {
	widgets := compact.NewCompact(t)
	return &TaskWidget{
		Task:    t,
		Widgets: widgets,
		display: true,
	}
}
