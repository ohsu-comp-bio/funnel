package datastore

import (
	"cloud.google.com/go/datastore"
	"github.com/ohsu-comp-bio/funnel/tes"
	"golang.org/x/net/context"
	"google.golang.org/api/iterator"
)

// GetTask implements the TES GetTask interface.
func (d *Datastore) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	key := taskKey(req.Id)

	var props datastore.PropertyList
	err := d.client.Get(ctx, key, &props)
	if err == datastore.ErrNoSuchEntity {
		return nil, tes.ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	task := &tes.Task{}
	if err := unmarshalTask(task, props); err != nil {
		return nil, err
	}

	switch req.View {
	case tes.View_MINIMAL.String():
		task = task.GetMinimalView()
	case tes.View_FULL.String():
		// Determine the keys needed to load the various parts of the full view.
		parts := viewPartKeys(task)
		err := d.getFullView(ctx, parts, map[string]*tes.Task{
			task.Id: task,
		})
		if err != nil {
			return nil, err
		}
	}

	return task, nil
}

// getFullView retrieve the various parts of the full view from the database
// and unmarshals those into their respective tasks. This handles unmarshaling
// multiple tasks in one call, in order to support ListTasks. The "tasks" arg
// is a map of task ID -> task.
func (d *Datastore) getFullView(ctx context.Context, keys []*datastore.Key, tasks map[string]*tes.Task) error {
	proplists := make([]datastore.PropertyList, len(keys))
	err := d.client.GetMulti(ctx, keys, proplists)
	merr, isMerr := err.(datastore.MultiError)
	if err != nil && !isMerr {
		return err
	}

	for i, props := range proplists {
		if merr != nil && merr[i] != nil {
			// The task doesn't have this part, e.g. it doesn't have stdout for an executor.
			// That's ok, skip this part.
			continue
		}
		id := keys[i].Parent.Name
		task := tasks[id]
		if err := unmarshalPart(task, props); err != nil {
			return err
		}
	}
	return nil
}

// ListTasks implements the TES ListTasks interface.
func (d *Datastore) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {

	page := req.PageToken
	size := tes.GetPageSize(req.GetPageSize())
	q := datastore.NewQuery("Task").KeysOnly().Limit(size).Order("-CreationTime")

	if page != "" {
		c, err := datastore.DecodeCursor(page)
		if err != nil {
			return nil, err
		}
		q = q.Start(c)
	}

	if req.State != tes.Unknown {
		q = q.FilterField("State", "=", int32(req.State))
	}

	for k, v := range req.GetTags() {
		q = q.FilterField("TagStrings", "=", encodeKV(k, v))
	}

	var tasks []*tes.Task
	var keys []*datastore.Key
	var parts []*datastore.Key
	byID := map[string]*tes.Task{}

	it := d.client.Run(ctx, q)
	for {
		key, err := it.Next(nil)
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	proplists := make([]datastore.PropertyList, len(keys))
	err := d.client.GetMulti(ctx, keys, proplists)
	if err != nil {
		return nil, err
	}

	for _, props := range proplists {
		task := &tes.Task{}
		if err := unmarshalTask(task, props); err != nil {
			return nil, err
		}

		switch req.View {
		case tes.View_MINIMAL.String():
			task = task.GetMinimalView()
		case tes.View_FULL.String():
			// Determine the keys needed to load the various parts of the full view.
			parts = append(parts, viewPartKeys(task)...)
			byID[task.Id] = task
		}
		tasks = append(tasks, task)
	}

	// Load the full view parts
	if req.View == tes.View_FULL.String() {
		err := d.getFullView(ctx, parts, byID)
		if err != nil {
			return nil, err
		}
	}

	resp := &tes.ListTasksResponse{Tasks: tasks}

	if len(keys) == size {
		c, err := it.Cursor()
		if err != nil {
			return nil, err
		}
		resp.NextPageToken = c.String()
	}

	return resp, nil
}
