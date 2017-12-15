package storage

import (
	"context"
	"github.com/cenkalti/backoff"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
)

type retrier struct {
	backend     Backend
	shouldRetry func(err error) bool
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

func (r *retrier) Get(ctx context.Context, url, path string, class tes.FileType) error {
	// Default backoff parameters are here:
	// https://github.com/cenkalti/backoff/blob/master/exponential.go#L74
	eb := backoff.NewExponentialBackOff()
	b := backoff.WithContext(eb, ctx)
	return backoff.Retry(func() error {
		err := r.backend.Get(ctx, url, path, class)
		return r.checkErr(err)
	}, b)
}

func (r *retrier) PutFile(ctx context.Context, url, path string) error {
	// Default backoff parameters are here:
	// https://github.com/cenkalti/backoff/blob/master/exponential.go#L74
	eb := backoff.NewExponentialBackOff()
	b := backoff.WithContext(eb, ctx)
	return backoff.Retry(func() error {
		err := r.backend.PutFile(ctx, url, path)
		return r.checkErr(err)
	}, b)
}

func (r *retrier) SupportsGet(url string, class tes.FileType) error {
	// Default backoff parameters are here:
	// https://github.com/cenkalti/backoff/blob/master/exponential.go#L74
	eb := backoff.NewExponentialBackOff()
	return backoff.Retry(func() error {
		err := r.backend.SupportsGet(url, class)
		return r.checkErr(err)
	}, eb)
}

func (r *retrier) SupportsPut(url string, class tes.FileType) error {
	// Default backoff parameters are here:
	// https://github.com/cenkalti/backoff/blob/master/exponential.go#L74
	eb := backoff.NewExponentialBackOff()
	return backoff.Retry(func() error {
		err := r.backend.SupportsPut(url, class)
		return r.checkErr(err)
	}, eb)
}
