package server

import (
	"github.com/boltdb/bolt"
	proto "github.com/golang/protobuf/proto"
	pbr "tes/server/proto"
)

func getWorker(tx *bolt.Tx, id string) (*pbr.Worker, error) {
	worker := &pbr.Worker{
		Id: id,
	}

	data := tx.Bucket(Workers).Get([]byte(id))
	if data != nil {
		proto.Unmarshal(data, worker)
	}

	if worker.Assigned == nil {
		worker.Assigned = map[string]bool{}
	}
	if worker.Active == nil {
		worker.Active = map[string]bool{}
	}
	return worker, nil
}

func putWorker(tx *bolt.Tx, worker *pbr.Worker) error {
	bw := tx.Bucket(Workers)
	data, _ := proto.Marshal(worker)
	bw.Put([]byte(worker.Id), data)
	return nil
}
