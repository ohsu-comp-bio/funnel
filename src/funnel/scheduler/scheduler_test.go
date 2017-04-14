package scheduler

import (
	"funnel/config"
	"funnel/logger"
	"funnel/proto/tes"
	"funnel/scheduler/mocks"
	"golang.org/x/net/context"
	"io/ioutil"
	"testing"
)

func init() {
	logger.ForceColors()
}

type BlankBackend struct{}

func (b *BlankBackend) Schedule(*tes.Job) *Offer {
	return nil
}

func TestBackendCaching(t *testing.T) {
	conf := config.DefaultConfig()
	conf.Scheduler = "test"
	f, _ := ioutil.TempDir("", "funnel-test-")
	conf.WorkDir = f

	db := &mocks.Database{}
	db.On("CheckWorkers").Return(nil)
	db.On("ReadQueue", 10).Return([]*tes.Job{})

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
