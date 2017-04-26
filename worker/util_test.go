package worker

import (
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

func addTask(tasks map[string]*pbf.TaskWrapper, t *tes.Task) {
	tasks[t.Id] = &pbf.TaskWrapper{Task: t}
}
