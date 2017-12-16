package storage

import (
	"context"
	"github.com/cenkalti/backoff"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"time"
)

type retrier struct {
	backend     Backend
	shouldRetry func(err error) bool
	maxTries    int
}

func (r *retrier) checkErr(err error) error {
	switch {
	case err != nil && r.shouldRetry != nil && !r.shouldRetry(err):
		return &backoff.PermanentError{Err: err}
	case err != nil:
		return err
	default:
		return nil
	}
}

func (r *retrier) backoff() backoff.BackOff {
	// Default backoff parameters are here:
	// https://github.com/cenkalti/backoff/blob/master/exponential.go#L74
	eb := backoff.NewExponentialBackOff()
	// Disable max retry time, since it is incompatible with upload time for large objects.
	eb.MaxElapsedTime = 0
	// Cap the number of retry attempts.
	return withMaxTries(eb, uint64(r.maxTries))
}

func (r *retrier) Get(ctx context.Context, url, path string, class tes.FileType) error {
	b := backoff.WithContext(r.backoff(), ctx)
	return backoff.Retry(func() error {
		err := r.backend.Get(ctx, url, path, class)
		return r.checkErr(err)
	}, b)
}

func (r *retrier) PutFile(ctx context.Context, url, path string) error {
	b := backoff.WithContext(r.backoff(), ctx)
	return backoff.Retry(func() error {
		err := r.backend.PutFile(ctx, url, path)
		return r.checkErr(err)
	}, b)
}

func (r *retrier) SupportsGet(url string, class tes.FileType) error {
	b := r.backoff()
	return backoff.Retry(func() error {
		err := r.backend.SupportsGet(url, class)
		return r.checkErr(err)
	}, b)
}

func (r *retrier) SupportsPut(url string, class tes.FileType) error {
	b := r.backoff()
	return backoff.Retry(func() error {
		err := r.backend.SupportsPut(url, class)
		return r.checkErr(err)
	}, b)
}

// copied because of an off-by-one bug
// https://github.com/cenkalti/backoff/blob/master/tries.go
// https://github.com/cenkalti/backoff/issues/54
func withMaxTries(b backoff.BackOff, max uint64) backoff.BackOff {
	return &backOffTries{delegate: b, maxTries: max}
}

type backOffTries struct {
	delegate backoff.BackOff
	maxTries uint64
	numTries uint64
}

func (b *backOffTries) NextBackOff() time.Duration {
	if b.maxTries > 0 {
		b.numTries++
		if b.numTries >= b.maxTries {
			return backoff.Stop
		}
	}
	return b.delegate.NextBackOff()
}

func (b *backOffTries) Reset() {
	b.numTries = 0
	b.delegate.Reset()
}
