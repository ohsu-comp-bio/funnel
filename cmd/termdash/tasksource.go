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
	List(bool, bool) TaskWidgets
	Get(string) *TaskWidget
	GetNextPage() string
	GetPreviousPage() string
}

type TaskSource struct {
	client   *client.Client
	pageSize uint32
	pPage    []string
	cPage    string
	nPage    string
	tasks    TaskWidgets
	lock     sync.RWMutex
}

func NewTaskSource(tesHTTPServerAddress string, pageSize uint32) *TaskSource {
	// init funnel http client
	cli := client.NewClient(tesHTTPServerAddress)
	ts := &TaskSource{
		client:   cli,
		pageSize: pageSize,
		lock:     sync.RWMutex{},
	}
	ts.tasks = ts.listTasks(false, false)
	return ts
}

func (ts *TaskSource) listTasks(previous, next bool) TaskWidgets {
	var tasks TaskWidgets

	if next && !previous {
		if ts.nPage != "" {
			if ts.cPage != "" {
				ts.pPage = append(ts.pPage, ts.cPage)
			}
			ts.cPage = ts.nPage
		}
	} else if previous && !next {
		if len(ts.pPage) > 0 {
			ts.cPage = ts.pPage[len(ts.pPage)-1]
			ts.pPage = ts.pPage[:len(ts.pPage)-1]
		} else {
			ts.cPage = ""
		}
	}

	resp, err := ts.client.ListTasks(context.Background(), &tes.ListTasksRequest{
		View:      tes.TaskView_BASIC,
		PageSize:  ts.pageSize,
		PageToken: ts.cPage,
	})
	if err != nil {
		header.SetError(fmt.Sprintf("%v", err))
		return nil
	} else {
		// header.SetError(fmt.Sprintf("Previous: %s; Next: %s; Current: %s", ts.pPage, ts.nPage, ts.cPage))
		header.SetError("")
	}
	ts.nPage = resp.NextPageToken

	for _, t := range resp.Tasks {
		tasks = append(tasks, NewTaskWidget(t))
	}

	return tasks
}

// Return array of tasks, sorted by field
func (ts *TaskSource) List(previous, next bool) TaskWidgets {
	ts.tasks = ts.listTasks(previous, next)

	ts.lock.Lock()
	var tasks TaskWidgets
	for _, t := range ts.tasks {
		tasks = append(tasks, t)
	}
	ts.lock.Unlock()

	sort.Sort(tasks)
	tasks.Filter()
	return tasks
}

// Get a single task, by ID
func (ts *TaskSource) Get(id string) *TaskWidget {
	defer func() {
		if r := recover(); r != nil {
			if header != nil {
				header.SetError(fmt.Sprintf("%v", r))
			}
		}
	}()

	task, err := ts.client.GetTask(context.Background(), &tes.GetTaskRequest{
		Id:   id,
		View: tes.TaskView_FULL,
	})
	if err != nil {
		panic(err)
	}
	return NewTaskWidget(task)
}

func (ts *TaskSource) GetNextPage() string {
	return ts.nPage
}

func (ts *TaskSource) GetPreviousPage() string {
	if len(ts.pPage) > 0 {
		return ts.pPage[len(ts.pPage)-1]
	} else if ts.cPage != "" {
		return ts.cPage
	}
	return ""
}
