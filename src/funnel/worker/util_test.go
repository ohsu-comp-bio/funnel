package worker

import (
	pbe "funnel/ga4gh"
	pbr "funnel/server/proto"
)

func addJob(jobs map[string]*pbr.JobWrapper, j *pbe.Job) {
	jobs[j.JobID] = &pbr.JobWrapper{Job: j}
}
