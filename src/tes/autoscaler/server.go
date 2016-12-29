package autoscaler

// TODO this is the start of a very basic scaling algorithm
//      if the queue is bigger than N, add workers
//      Things to think about:
//      - how does scaling take resource requirements into consideration
//      - how do does scaler switch modes so that it's not checking scale
//        while new workers are starting?
//      - how does the scaling algorithm become reusable across multiple
//        backends?
//      - how does scaler detect that the condor queue is full?
//        i.e. there is no room to scale up.

import (
	"tes/scheduler"
	"time"
)

func PollAutoscaler(autoscaler Autoscaler, d time.Duration) {
	ticker := time.NewTicker(d)
	for _ = range ticker.C {
		autoscaler.Analyze()
	}
}

type Autoscaler interface {
	Analyze()
}

func DumbLocalAutoscaler(schedAddr string) Autoscaler {
	sched, _ := scheduler.NewClient(schedAddr)
	factory := LocalWorkerFactory{}
	return &DumbAutoscaler{sched, factory}
}

func DumbCondorAutoscaler(schedAddr string, proxyAddr string) Autoscaler {
	sched, _ := scheduler.NewClient(schedAddr)
	client, _ := NewCondorProxyClient(proxyAddr)
	factory := CondorWorkerFactory{schedAddr, client}
	return &DumbAutoscaler{sched, factory}
}
