package worker

import (
	"errors"
	"github.com/ohsu-comp-bio/funnel/config"
	sched_mocks "github.com/ohsu-comp-bio/funnel/scheduler/mocks"
	"github.com/stretchr/testify/mock"
	"testing"
	"time"
)

// Test calling Worker.Stop()
func TestStopWorker(t *testing.T) {
	w, _ := NewWorker(config.DefaultConfig().Worker)
	done := make(chan struct{})
	go func() {
		w.Run()
		close(done)
	}()
	timeout := time.NewTimer(time.Millisecond * 4)
	w.Stop()

	// Wait for either the worker to be done, or the test to timeout
	select {
	case <-timeout.C:
		t.Error("Expected worker to be done")
	case <-done:
		// Worker is done
	}
}

// Mainly exercising a panic bug caused by an unhandled
// error from client.GetWorker().
func TestGetWorkerFail(t *testing.T) {
	// Create worker
	conf := config.DefaultConfig().Worker
	w, err := NewWorker(conf)
	if err != nil {
		t.Error(err)
	}

	// Override worker client with new mock
	m := new(sched_mocks.Client)
	s := &schedClient{m, conf}
	w.sched = s

	// Set GetWorker to return an error
	m.On("GetWorker", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("TEST"))
	w.Sync()
}
