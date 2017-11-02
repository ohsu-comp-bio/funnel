package termdash

import (
	"fmt"
	"github.com/ohsu-comp-bio/funnel/client"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"golang.org/x/net/context"
	"sort"
	"sync"
)

type TesTaskSource interface {
	All() TaskWidgets
	Get(string) *TaskWidget
}

type TaskSource struct {
	client *client.Client
	tasks  TaskWidgets
	lock   sync.RWMutex
}

func NewTaskSource(tesHTTPServerAddress string) *TaskSource {
	// init funnel http client
	cli := client.NewClient(tesHTTPServerAddress)
	cm := &TaskSource{
		client: cli,
		lock:   sync.RWMutex{},
	}
	cm.tasks = cm.listTasks()
	return cm
}

func (cm *TaskSource) listTasks() TaskWidgets {
	var tasks TaskWidgets
	var page string

	defer func() {
		if r := recover(); r != nil {
			if header != nil {
				header.SetError(fmt.Sprintf("%v", r))
			}
		}
	}()

	for {
		resp, err := cm.client.ListTasks(context.Background(), &tes.ListTasksRequest{
			View:      tes.TaskView_BASIC,
			PageToken: page,
		})
		if err != nil {
			panic(err)
		} else {
			header.SetError("")
		}
		for _, t := range resp.Tasks {
			tasks = append(tasks, NewTaskWidget(t))
		}
		page = resp.NextPageToken
		if page == "" {
			break
		}
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
	defer func() {
		if r := recover(); r != nil {
			if header != nil {
				header.SetError(fmt.Sprintf("%v", r))
			}
		}
	}()

	task, err := cm.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   id,
		View: tes.TaskView_FULL,
	})
	if err != nil {
		panic(err)
	}
	return NewTaskWidget(task)
}
