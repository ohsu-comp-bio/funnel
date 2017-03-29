package worker

import (
	pbf "funnel/proto/funnel"
	tes "funnel/proto/tes"
)

func addJob(jobs map[string]*pbf.JobWrapper, j *tes.Job) {
	jobs[j.JobID] = &pbf.JobWrapper{Job: j}
}
