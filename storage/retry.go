package storage

import (
	"context"

	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/util"
)

type retrier struct {
	*util.Retrier
	backend Backend
}

func (r *retrier) Get(ctx context.Context, url, path string, class tes.FileType) error {
	return r.Retry(ctx, func() error {
		return r.backend.Get(ctx, url, path, class)
	})
}

func (r *retrier) PutFile(ctx context.Context, url, path string) error {
	return r.Retry(ctx, func() error {
		return r.backend.PutFile(ctx, url, path)
	})
}

func (r *retrier) SupportsGet(url string, class tes.FileType) error {
	return r.Retry(context.Background(), func() error {
		return r.backend.SupportsGet(url, class)
	})
}

func (r *retrier) SupportsPut(url string, class tes.FileType) error {
	return r.Retry(context.Background(), func() error {
		return r.backend.SupportsPut(url, class)
	})
}
