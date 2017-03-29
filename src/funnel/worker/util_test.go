package worker

import (
	tes "funnel/proto/tes"
	pbf "funnel/proto/funnel"
)

func addJob(jobs map[string]*pbf.JobWrapper, j *tes.Job) {
	jobs[j.JobID] = &pbf.JobWrapper{Job: j}
}
