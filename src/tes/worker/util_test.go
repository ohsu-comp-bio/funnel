package worker

import (
	"tes/config"
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
)

func addJob(jobs map[string]*pbr.JobWrapper, j *pbe.Job) {
	jobs[j.JobID] = &pbr.JobWrapper{Job: j}
}

func noopRunJob(l JobControl, c config.Worker, j *pbr.JobWrapper, u logUpdateChan) {
}

type mockRunner struct {
	ctrl    JobControl
	conf    config.Worker
	wrapper *pbr.JobWrapper
	updates logUpdateChan
}
