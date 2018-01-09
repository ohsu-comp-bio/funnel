package rpc

import (
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {

	eb := newExponentialBackoff()

	getMinMax := func(val time.Duration, randomizationFactor float64) (min time.Duration, max time.Duration) {
		v := float64(val)
		delta := randomizationFactor * v
		min = time.Duration(v - delta)
		max = time.Duration(v + delta)
		return
	}

	maxBackoffMin, maxBackoffMax := getMinMax(eb.Max, eb.RandomizationFactor)
	t.Logf("maxBackoff jittered interval [%v, %v]", maxBackoffMin, maxBackoffMax)

	for i := 0; i < 10; i++ {
		attempt := uint(i)
		t.Log("attempt", attempt)

		nb := eb.Backoff(attempt)
		nbj := eb.BackoffWithJitter(attempt)
		min, max := getMinMax(nb, eb.RandomizationFactor)

		if nb > eb.Max {
			if nbj < maxBackoffMin || nbj > maxBackoffMax {
				t.Errorf("expected backoff to be limited by maxBackoff parameter. Got %v", nbj)
			}
		} else {
			if nbj < min || nbj > max {
				t.Errorf("expected backoff to be in interval [%v, %v]. Got %v", min, max, nbj)
			}
		}
	}
}
