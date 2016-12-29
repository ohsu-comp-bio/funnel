package main

import (
	"flag"
	"tes/autoscaler"
	"time"
)

func main() {
	schedArg := flag.String("sched", "localhost:9090", "Address of the scheduler, e.g. localhost:9090")
	proxyArg := flag.String("proxy", "localhost:9054", "Address of the Condor proxy, e.g. localhost:9054")
	flag.Parse()
	scaler := autoscaler.DumbCondorAutoscaler(*schedArg, *proxyArg)
	autoscaler.PollAutoscaler(scaler, time.Second*5)
}
