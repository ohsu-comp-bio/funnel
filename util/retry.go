package util

import (
	"context"
	"time"

	"github.com/cenkalti/backoff"
)

type MaxRetrier struct {
	Backoff     backoff.BackOff
	ShouldRetry func(err error) bool
	MaxTries    int
}

func (r *MaxRetrier) Retry(ctx context.Context, f func() error) error {
	b := backoff.WithContext(r.backoff(), ctx)
	return backoff.Retry(func() error {
		err := f()
		return r.checkErr(err)
	}, b)
}

func (r *MaxRetrier) checkErr(err error) error {
	switch {
	case err != nil && r.ShouldRetry != nil && !r.ShouldRetry(err):
		return &backoff.PermanentError{Err: err}
	case err != nil:
		return err
	default:
		return nil
	}
}

func (r *MaxRetrier) backoff() backoff.BackOff {
	// default for Backoff if unset
	if r.Backoff == nil {
		// Default backoff parameters are here:
		// https://github.com/cenkalti/backoff/blob/master/exponential.go#L74
		eb := backoff.NewExponentialBackOff()
		eb.InitialInterval = time.Millisecond * 100
		eb.MaxInterval = time.Second * 60
		eb.Multiplier = 1.5
		// Disable max retry time, since it is incompatible with upload time for large objects.
		eb.MaxElapsedTime = 0
		r.Backoff = eb
	}
	// default for MaxTries if unset
	if r.MaxTries == 0 {
		r.MaxTries = 10
	}

	// Cap the number of retry attempts.
	return backoff.WithMaxRetries(r.Backoff, uint64(r.MaxTries-1))
}
