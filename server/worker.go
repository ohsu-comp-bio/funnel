package server

import (
	"bytes"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	pbf "github.com/ohsu-comp-bio/funnel/proto/funnel"
)

func getWorker(tx *bolt.Tx, id string) *pbf.Worker {
	worker := &pbf.Worker{
		Id: id,
	}

	data := tx.Bucket(Workers).Get([]byte(id))
	if data != nil {
		proto.Unmarshal(data, worker)
	}

	worker.Tasks = map[string]*pbf.TaskWrapper{}
	// Prefix scan for keys that start with worker ID
	c := tx.Bucket(WorkerTasks).Cursor()
	prefix := []byte(id)
	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		taskID := string(v)
		task := getTask(tx, taskID)
		auth := getTaskAuth(tx, taskID)
		wrapper := &pbf.TaskWrapper{
			Task: task,
			Auth: auth,
		}
		worker.Tasks[taskID] = wrapper
	}

	if worker.Metadata == nil {
		worker.Metadata = map[string]string{}
	}

	return worker
}

func putWorker(tx *bolt.Tx, worker *pbf.Worker) {
	// Tasks are not saved in the database under the worker,
	// they are stored in a separate bucket and linked via an index.
	// The same protobuf message is used for both communication and database,
	// so we have to set nil here.
	//
	// Also, this modifies the worker, so copy it first.
	w := proto.Clone(worker).(*pbf.Worker)
	w.Tasks = nil
	data, _ := proto.Marshal(w)
	tx.Bucket(Workers).Put([]byte(w.Id), data)
}
