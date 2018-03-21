package termdash

import (
	"github.com/ohsu-comp-bio/funnel/cmd/termdash/compact"
	"github.com/ohsu-comp-bio/funnel/tes"
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
