package main

import (
	"tes/autoscaler"
	"time"
)

func main() {
	scaler := autoscaler.DumbLocalAutoscaler("localhost:9090")
	autoscaler.PollAutoscaler(scaler, time.Second*5)
}
