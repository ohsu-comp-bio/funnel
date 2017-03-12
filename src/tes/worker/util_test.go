package worker

import (
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
)

func addJob(jobs map[string]*pbr.JobWrapper, j *pbe.Job) {
	jobs[j.JobID] = &pbr.JobWrapper{Job: j}
}
