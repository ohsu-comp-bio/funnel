package scheduler

import (
	"github.com/ohsu-comp-bio/funnel/config"
	"github.com/ohsu-comp-bio/funnel/logger"
	"github.com/ohsu-comp-bio/funnel/proto/tes"
	"github.com/ohsu-comp-bio/funnel/scheduler/mocks"
	"golang.org/x/net/context"
	"io/ioutil"
	"testing"
)

func init() {
	logger.Configure(logger.DebugConfig())
}

type BlankBackend struct{}

func (b *BlankBackend) Schedule(*tes.Task) *Offer {
	return nil
}

func TestBackendCaching(t *testing.T) {
	conf := config.DefaultConfig()
	conf.Scheduler = "test"
	f, _ := ioutil.TempDir("", "funnel-test-")
	conf.WorkDir = f

	db := &mocks.Database{}
	db.On("CheckWorkers").Return(nil)
	db.On("ReadQueue", 10).Return([]*tes.Task{})

	s, _ := NewScheduler(db, conf)
	calls := 0
	s.AddBackend(&BackendPlugin{
		Name: "test",
		Create: func(config.Config) (Backend, error) {
			calls++
			return &BlankBackend{}, nil
		},
	})
	ctx := context.Background()
	s.Schedule(ctx)
	s.Schedule(ctx)
	if calls != 1 {
		logger.Debug("Calls", calls)
		t.Error("Expected one call")
		return
	}
}
