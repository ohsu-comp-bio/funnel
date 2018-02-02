package util

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
)

func TestMaxRetrier(t *testing.T) {
	r := &MaxRetrier{
		MaxTries:    3,
		ShouldRetry: nil,
		Backoff: &backoff.ExponentialBackOff{
			InitialInterval:     time.Millisecond * 10,
			MaxInterval:         time.Second * 60,
			Multiplier:          2.0,
			MaxElapsedTime:      0,
			RandomizationFactor: 0,
			Clock:               backoff.SystemClock,
		},
	}
	bg := context.Background()

	i := 0
	r.Retry(bg, func() error {
		i += 1
		return fmt.Errorf("always error")
	})
	if i != 3 {
		t.Error("unexpected number of retries", i)
	}
	next := r.Backoff.NextBackOff()
	if next != time.Millisecond*40 {
		t.Error("unexpected next backoff", next)
	}

	r.Retry(bg, func() error {
		return nil
	})
	next = r.Backoff.NextBackOff()
	if next != time.Millisecond*10 {
		t.Error("unexpected next backoff", next)
	}
}
