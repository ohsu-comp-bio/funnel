package gce

import (
	pbr "tes/server/proto"
	"testing"
)

type mockTracker struct{}

func (mt *mockTracker) Workers() []*pbr.Worker {
	return []*pbr.Worker{}
}

func TestBasic(t *testing.T) {
	s := scheduler{
		conf:    conf,
		tracker: &mockTracker{},
	}
	j := &pbe.Job{
		JobID: "test-job-1",
		Task:  &pbe.Task{},
	}
	o := s.Schedule(j)
	if o != nil {
		t.Error("Expected nil offer")
	}
}
