package ga4gh_engine_scaling

import (
	"ga4gh-server/proto"
	"ga4gh-tasks"
)

type Scaler interface {
	JobAdded(*ga4gh_task_exec.Resources)
	PingReceived(*ga4gh_task_ref.WorkerInfo)
}

type ScalerInit func(map[string]string) Scaler

var ScalingMethods = map[string]ScalerInit{
	"local": NewLocalScaler,
}

type LocalScaler struct {
}

func NewLocalScaler(config map[string]string) Scaler {
	return LocalScaler{}
}

func (self LocalScaler) JobAdded(request *ga4gh_task_exec.Resources) {
	//check for number of running workers

	//launch new worker if needed and possible
}

func (self LocalScaler) PingReceived(worker *ga4gh_task_ref.WorkerInfo) {
	//do something here
}
