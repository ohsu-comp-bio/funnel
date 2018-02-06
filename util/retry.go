package util

import (
	"context"
	"time"

	"github.com/cenkalti/backoff"
)

type Retrier struct {
	InitialInterval     time.Duration
	MaxInterval         time.Duration
	Multiplier          float64
	RandomizationFactor float64
	MaxElapsedTime      time.Duration
	MaxTries            int
	ShouldRetry         func(err error) bool
	backoff             backoff.BackOff
}

func NewRetrier() *Retrier {
	// based on https://github.com/cenkalti/backoff/blob/master/exponential.go#L74
	return &Retrier{
		InitialInterval:     time.Millisecond * 500,
		MaxInterval:         time.Second * 60,
		Multiplier:          1.5,
		RandomizationFactor: 0.5,
		MaxElapsedTime:      time.Minute * 15,
		MaxTries:            10,
		ShouldRetry:         nil,
	}
}

func (r *Retrier) Retry(ctx context.Context, f func() error) error {
	b := backoff.WithContext(r.withTries(), ctx)
	return backoff.Retry(func() error { return r.checkErr(f()) }, b)
}

func (r *Retrier) checkErr(err error) error {
	switch {
	case err != nil && r.ShouldRetry != nil && !r.ShouldRetry(err):
		return &backoff.PermanentError{Err: err}
	case err != nil:
		return err
	default:
		return nil
	}
}

func (r *Retrier) withTries() backoff.BackOff {
	r.backoff = &backoff.ExponentialBackOff{
		InitialInterval:     r.InitialInterval,
		MaxInterval:         r.MaxInterval,
		Multiplier:          r.Multiplier,
		RandomizationFactor: r.RandomizationFactor,
		MaxElapsedTime:      r.MaxElapsedTime,
		Clock:               backoff.SystemClock,
	}

	max := r.MaxTries - 1
	if max < 0 {
		max = 0
	}

	// Cap the number of retry attempts.
	return backoff.WithMaxRetries(r.backoff, uint64(max))
}
