package mocks

import context "golang.org/x/net/context"
import ga4gh_task_exec "tes/ga4gh"
import ga4gh_task_ref "tes/server/proto"
import mock "github.com/stretchr/testify/mock"

// Database is an autogenerated mock type for the Database type
type Database struct {
	mock.Mock
}

// AssignJob provides a mock function with given fields: _a0, _a1
func (_m *Database) AssignJob(_a0 *ga4gh_task_exec.Job, _a1 *ga4gh_task_ref.Worker) {
	_m.Called(_a0, _a1)
}

// CheckWorkers provides a mock function with given fields:
func (_m *Database) CheckWorkers() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetWorkers provides a mock function with given fields: _a0, _a1
func (_m *Database) GetWorkers(_a0 context.Context, _a1 *ga4gh_task_ref.GetWorkersRequest) (*ga4gh_task_ref.GetWorkersResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *ga4gh_task_ref.GetWorkersResponse
	if rf, ok := ret.Get(0).(func(context.Context, *ga4gh_task_ref.GetWorkersRequest) *ga4gh_task_ref.GetWorkersResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ga4gh_task_ref.GetWorkersResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *ga4gh_task_ref.GetWorkersRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadQueue provides a mock function with given fields: n
func (_m *Database) ReadQueue(n int) []*ga4gh_task_exec.Job {
	ret := _m.Called(n)

	var r0 []*ga4gh_task_exec.Job
	if rf, ok := ret.Get(0).(func(int) []*ga4gh_task_exec.Job); ok {
		r0 = rf(n)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*ga4gh_task_exec.Job)
		}
	}

	return r0
}

// UpdateWorker provides a mock function with given fields: _a0, _a1
func (_m *Database) UpdateWorker(_a0 context.Context, _a1 *ga4gh_task_ref.Worker) (*ga4gh_task_ref.UpdateWorkerResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *ga4gh_task_ref.UpdateWorkerResponse
	if rf, ok := ret.Get(0).(func(context.Context, *ga4gh_task_ref.Worker) *ga4gh_task_ref.UpdateWorkerResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ga4gh_task_ref.UpdateWorkerResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *ga4gh_task_ref.Worker) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
