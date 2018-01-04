package rpc

import (
	"math"
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	var initialBackoff = 5 * time.Second
	var maxBackoff = 1 * time.Minute
	var multiplier = 2.0
	var randomizationFactor = 0.5

	getMinMax := func(val float64) (min float64, max float64) {
		delta := randomizationFactor * val
		min = val - delta
		max = val + delta
		return
	}

	maxBackoffMin, maxBackoffMax := getMinMax(float64(maxBackoff))
	t.Logf("maxBackoff jittered interval [%v, %v]", maxBackoffMin, maxBackoffMax)

	for i := 0; i < 10; i++ {
		attempt := uint(i)
		t.Log("attempt", attempt)

		min, max := getMinMax(float64(initialBackoff) * math.Pow(multiplier, float64(attempt)))
		b := exponentialBackoff(attempt)

		if min > float64(maxBackoff) {
			if float64(b) < maxBackoffMin || float64(b) > maxBackoffMax {
				t.Errorf("expected backoff to be limited by maxBackoff parameter. Got %v", float64(b))
			}
		} else {
			if float64(b) < min || float64(b) > max {
				t.Errorf("expected backoff to be in interval [%v, %v]. Got %v", min, max, float64(b))
			}
		}
	}
}
