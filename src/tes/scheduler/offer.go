package scheduler

import (
	"tes/config"
	pbe "tes/ga4gh"
)

// Resources describe a set of computational resources, e.g. CPU, RAM, etc.
//
// The types are very specific because they match the protobuf
// Resources message types.
type Resources struct {
	// TODO doesn't say anything about size of CPU.
	//      does that matter these days?
	CPUs uint32
	RAM  float64
	//Disk float64
}

type Scores struct {
	CPUUsage       float32
	CPUUsageWeight float32
	RAMUsage       float32
	RAMUsageWeight float32
	// TODO ideas
	// - startup time: could account for VM initialization
	// - cost: could account for cloud compute costs
	// - allow scores to track data over time and collect time-based metrics
}

type Worker struct {
	ID        string
	Resources Resources
	Available Resources
	Conf      config.Worker
	Scores    Scores
}

func (w *Worker) CalcScores(j *pbe.Job) {
	req := j.Task.Resources
	tot := w.Resources
	avail := w.Available
	s := &w.Scores

	s.CPUUsage = float32(avail.CPUs+req.MinimumCpuCores) / float32(tot.CPUs)
	s.RAMUsage = float32(avail.RAM + req.MinimumRamGb/tot.RAM)
}

func (w *Worker) Score() float32 {
	s := &w.Scores
	return s.CPUUsage*s.CPUUsageWeight +
		s.RAMUsage*s.RAMUsageWeight
}

// ByScore is a helper which implements Go's sort.Interface to sort
// workers by their score (i.e. worker.Score())
type ByScore []*Worker

func (s ByScore) Len() int {
	return len(s)
}
func (s ByScore) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByScore) Less(i, j int) bool {
	return s[i].Score() < s[j].Score()
}
