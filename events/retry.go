package events

import (
	"context"

	"github.com/ohsu-comp-bio/funnel/util"
)

type Retrier struct {
	*util.MaxRetrier
	Writer Writer
}

func (r *Retrier) WriteEvent(ctx context.Context, e *Event) error {
	return r.Retry(ctx, func() error {
		return r.Writer.WriteEvent(ctx, e)
	})
}
