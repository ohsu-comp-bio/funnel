package server

import (
	"bytes"
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	pbr "tes/server/proto"
)

func getWorker(tx *bolt.Tx, id string) *pbr.Worker {
	worker := &pbr.Worker{
		Id: id,
	}

	data := tx.Bucket(Workers).Get([]byte(id))
	if data != nil {
		proto.Unmarshal(data, worker)
	}

	worker.Jobs = map[string]*pbr.JobWrapper{}
	// Prefix scan for keys that start with worker ID
	c := tx.Bucket(WorkerJobs).Cursor()
	prefix := []byte(id)
	for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
		jobID := string(v)
		job := getJob(tx, jobID)
		auth := getJobAuth(tx, jobID)
		wrapper := &pbr.JobWrapper{
			Job:  job,
			Auth: auth,
		}
		worker.Jobs[jobID] = wrapper
	}
	return worker
}

func putWorker(tx *bolt.Tx, worker *pbr.Worker) {
	// Jobs are not saved in the database under the worker,
	// they are stored in a separate bucket and linked via an index.
	// The same protobuf message is used for both communication and database,
	// so we have to set nil here.
	//
	// Also, this modifies the worker, so copy it first.
	w := proto.Clone(worker).(*pbr.Worker)
	w.Jobs = nil
	data, _ := proto.Marshal(w)
	tx.Bucket(Workers).Put([]byte(w.Id), data)
}
