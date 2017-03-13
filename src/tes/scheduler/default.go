package scheduler

import (
	"tes/config"
	pbe "tes/ga4gh"
	pbr "tes/server/proto"
)

// DefaultScheduleAlgorithm implements a simple scheduling algorithm
// that is (currently) common across a few scheduler backends.
// Given a job, list of workers, and weights, it returns the best Offer or nil.
func DefaultScheduleAlgorithm(j *pbe.Job, workers []*pbr.Worker, weights config.Weights) *Offer {

	offers := []*Offer{}
	for _, w := range workers {
		// Filter out workers that don't match the job request.
		// Checks CPU, RAM, disk space, ports, etc.
		if !Match(w, j, DefaultPredicates) {
			continue
		}

		sc := DefaultScores(w, j)
		sc = sc.Weighted(weights)

		offer := NewOffer(w, j, sc)
		offers = append(offers, offer)
	}

	// No matching workers were found.
	if len(offers) == 0 {
		return nil
	}

	SortByAverageScore(offers)
	return offers[0]
}
