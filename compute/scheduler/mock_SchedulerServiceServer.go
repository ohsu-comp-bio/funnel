// Code generated by mockery v1.0.0. DO NOT EDIT.
package scheduler

import (
	mock "github.com/stretchr/testify/mock"
	context "golang.org/x/net/context"
)

// MockSchedulerServiceServer is an autogenerated mock type for the SchedulerServiceServer type
type MockSchedulerServiceServer struct {
	UnimplementedSchedulerServiceServer
	mock.Mock
}

// DeleteNode provides a mock function with given fields: _a0, _a1
func (_m *MockSchedulerServiceServer) DeleteNode(_a0 context.Context, _a1 *Node) (*DeleteNodeResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *DeleteNodeResponse
	if rf, ok := ret.Get(0).(func(context.Context, *Node) *DeleteNodeResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*DeleteNodeResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *Node) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNode provides a mock function with given fields: _a0, _a1
func (_m *MockSchedulerServiceServer) GetNode(_a0 context.Context, _a1 *GetNodeRequest) (*Node, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *Node
	if rf, ok := ret.Get(0).(func(context.Context, *GetNodeRequest) *Node); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *GetNodeRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListNodes provides a mock function with given fields: _a0, _a1
func (_m *MockSchedulerServiceServer) ListNodes(_a0 context.Context, _a1 *ListNodesRequest) (*ListNodesResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *ListNodesResponse
	if rf, ok := ret.Get(0).(func(context.Context, *ListNodesRequest) *ListNodesResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ListNodesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *ListNodesRequest) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PutNode provides a mock function with given fields: _a0, _a1
func (_m *MockSchedulerServiceServer) PutNode(_a0 context.Context, _a1 *Node) (*PutNodeResponse, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *PutNodeResponse
	if rf, ok := ret.Get(0).(func(context.Context, *Node) *PutNodeResponse); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*PutNodeResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *Node) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
