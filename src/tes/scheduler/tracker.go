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

// NewTracker returns a new Tracker instance.
func NewTracker(conf config.Worker) *Tracker {
	return &Tracker{
		conf:    conf,
		workers: []*pbr.Worker{},
	}
}

// Tracker helps a scheduler track the state of workers by polling the server
// for updates. The scheduler uses the Workers() method to get the latest
// response.
type Tracker struct {
	conf    config.Worker
	workers []*pbr.Worker
	mtx     sync.Mutex
}

// Run starts polling the server with calls to GetWorkers()
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

// Workers returns the most recent set of workers received from the server.
func (t *Tracker) Workers() []*pbr.Worker {
	t.mtx.Lock()
	defer t.mtx.Unlock()

	return t.workers
}
