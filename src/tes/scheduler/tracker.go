package scheduler

import (
	"context"
	"sync"
	"tes/config"
	pbr "tes/server/proto"
	"time"
)

// tracker helps poll the database for updated worker information.
// TODO consider making an interface for this, which a condor tracker would implement
func NewTracker(conf config.Worker) *Tracker {
	return &Tracker{
		conf:    conf,
		workers: []*pbr.Worker{},
	}
}

type Tracker struct {
	conf    config.Worker
	workers []*pbr.Worker
	mtx     sync.Mutex
}

func (t *Tracker) Run() {
	client, _ := NewClient(t.conf)
	defer client.Close()

	ticker := time.NewTicker(t.conf.TrackerRate)

	for {
		<-ticker.C
		resp, err := client.GetWorkers(context.Background(), &pbr.GetWorkersRequest{})
		if err != nil {
			log.Error("Failed GetWorkers request. Recovering.", err)
			continue
		}
		t.mtx.Lock()
		t.workers = resp.Workers
		t.mtx.Unlock()
	}
}

func (t *Tracker) Workers() []*pbr.Worker {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	return t.workers
}
