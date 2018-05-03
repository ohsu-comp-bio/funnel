package storage

import (
	"context"
	"sync"

	"github.com/ohsu-comp-bio/funnel/util/fsutil"
)

// Transfer defines the interface of a single storage upload or download request.
//
// Transfer events (started, failed, finished, etc) are communicated
// via the Transfer interface.
type Transfer interface {
	URL() string
	Path() string
	Started()
	Finished(obj *Object)
	Failed(err error)
}

// Download downloads a list of transfers from storage, in parallel.
//
// Transfer events (started, failed, finished, etc) are communicated
// via the Transfer interface.
func Download(ctx context.Context, store Storage, transfers []Transfer) {
	wg := &sync.WaitGroup{}
	wg.Add(len(transfers))

	for _, x := range transfers {
		go func(x Transfer) {
			defer wg.Done()
			x.Started()

			var obj *Object
			err := fsutil.EnsurePath(x.Path())
			if err == nil {
				obj, err = store.Get(ctx, x.URL(), x.Path())
			}

			if err != nil {
				x.Failed(err)
			} else {
				x.Finished(obj)
			}
		}(x)
	}
	wg.Wait()
}

// Upload uploads a list of transfers to storage, in parallel.
//
// Transfer events (started, failed, finished, etc) are communicated
// via the Transfer interface.
func Upload(ctx context.Context, store Storage, transfers []Transfer) {
	wg := &sync.WaitGroup{}
	wg.Add(len(transfers))

	for _, x := range transfers {
		go func(x Transfer) {
			defer wg.Done()

			x.Started()
			obj, err := store.Put(ctx, x.URL(), x.Path())

			if err != nil {
				x.Failed(err)
			} else {
				x.Finished(obj)
			}
		}(x)
	}
	wg.Wait()
}
