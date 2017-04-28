package termdash

import (
	"github.com/ohsu-comp-bio/funnel/cmd/task"
	"sort"
	"sync"
)

type TesTaskSource interface {
	All() TaskWidgets
	Get(string) *TaskWidget
}

type TaskSource struct {
	client *task.Client
	tasks  TaskWidgets
	lock   sync.RWMutex
}

func NewTaskSource(tesHTTPServerAddress string) *TaskSource {
	// init funnel http client
	client := task.NewClient(tesHTTPServerAddress)
	cm := &TaskSource{
		client: client,
		lock:   sync.RWMutex{},
	}
	cm.tasks = cm.listTasks()
	return cm
}

func (cm *TaskSource) listTasks() TaskWidgets {
	resp, err := cm.client.ListTasks()
	if err != nil {
		panic(err)
	}
	var tasks TaskWidgets
	for _, t := range resp.Tasks {
		tasks = append(tasks, NewTaskWidget(t))
	}
	return tasks
}

// Return array of all tasks, sorted by field
func (cm *TaskSource) All() TaskWidgets {
	cm.tasks = cm.listTasks()

	cm.lock.Lock()
	var tasks TaskWidgets
	for _, t := range cm.tasks {
		tasks = append(tasks, t)
	}
	cm.lock.Unlock()

	sort.Sort(tasks)
	tasks.Filter()
	return tasks
}

// Get a single task, by ID
func (cm *TaskSource) Get(id string) *TaskWidget {
	task, err := cm.client.GetTask(id)
	if err != nil {
		panic(err)
	}
	return NewTaskWidget(task)
}
