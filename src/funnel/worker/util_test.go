package worker

import (
	pbf "funnel/proto/funnel"
	"funnel/proto/tes"
)

func addTask(tasks map[string]*pbf.TaskWrapper, t *tes.Task) {
	tasks[t.Id] = &pbf.TaskWrapper{Task: t}
}
