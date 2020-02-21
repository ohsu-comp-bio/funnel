package badger

import (
	"context"
	"fmt"

	badger "github.com/dgraph-io/badger/v2"
	proto "github.com/golang/protobuf/proto"
	"github.com/ohsu-comp-bio/funnel/tes"
)

// GetTask gets a task, which describes a running task
func (db *Badger) GetTask(ctx context.Context, req *tes.GetTaskRequest) (*tes.Task, error) {
	var task *tes.Task

	err := db.db.View(func(txn *badger.Txn) error {
		t, err := db.getTask(txn, req.Id)
		task = t
		return err
	})
	if err != nil {
		return nil, err
	}

	switch req.View {
	case tes.Minimal:
		task = task.GetMinimalView()
	case tes.Basic:
		task = task.GetBasicView()
	}
	return task, nil
}

// ListTasks returns a list of tasks.
func (db *Badger) ListTasks(ctx context.Context, req *tes.ListTasksRequest) (*tes.ListTasksResponse, error) {
	var tasks []*tes.Task
	pageSize := tes.GetPageSize(req.GetPageSize())

	err := db.db.View(func(txn *badger.Txn) error {

		it := txn.NewIterator(badger.IteratorOptions{
			// Keys (task IDs) are in ascending order, and we want the first page
			// to be the most recent task, so that's at the end of the list.
			Reverse:        true,
			PrefetchValues: true,
			PrefetchSize:   pageSize,
		})
		defer it.Close()

		i := 0

		// For pagination, figure out the starting key.
		if req.PageToken != "" {
			it.Seek(taskKey(req.PageToken))
			// Seek moves to the key, but the start of the page is the next key.
			it.Next()
		} else {
			it.Rewind()
		}

	taskLoop:
		for ; it.Valid() && len(tasks) < pageSize; it.Next() {
			var val []byte
			err := it.Item().Value(func(d []byte) error {
				val = copyBytes(d)
				return nil
			})
			if err != nil {
				return fmt.Errorf("loading item value: %s", err)
			}

			// Load task.
			task := &tes.Task{}
			err = proto.Unmarshal(val, task)
			if err != nil {
				return fmt.Errorf("unmarshaling data: %s", err)
			}

			// Filter tasks by tag.
			for k, v := range req.Tags {
				tval, ok := task.Tags[k]
				if !ok || tval != v {
					continue taskLoop
				}
			}

			// Filter tasks by state.
			if req.State != tes.Unknown && req.State != task.State {
				continue taskLoop
			}

			switch req.View {
			case tes.Minimal:
				task = task.GetMinimalView()
			case tes.Basic:
				task = task.GetBasicView()
			}

			tasks = append(tasks, task)
			i++
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	out := tes.ListTasksResponse{
		Tasks: tasks,
	}

	if len(tasks) == pageSize {
		out.NextPageToken = tasks[len(tasks)-1].Id
	}

	return &out, nil
}

func (db *Badger) getTask(txn *badger.Txn, id string) (*tes.Task, error) {
	item, err := txn.Get(taskKey(id))
	if err == badger.ErrKeyNotFound {
		return nil, tes.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("loading item: %s", err)
	}

	var val []byte
	err = item.Value(func(d []byte) error {
		val = copyBytes(d)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("loading item value: %s", err)
	}

	task := &tes.Task{}
	err = proto.Unmarshal(val, task)
	if err != nil {
		return nil, fmt.Errorf("unmarshaling data: %s", err)
	}
	return task, nil
}

func copyBytes(in []byte) []byte {
	out := make([]byte, len(in))
	copy(out, in)
	return out
}
