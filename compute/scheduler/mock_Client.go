// Code generated by mockery v1.0.0. DO NOT EDIT.
package scheduler

import (
	events "github.com/ohsu-comp-bio/funnel/events"
	context "golang.org/x/net/context"

	grpc "google.golang.org/grpc"

	mock "github.com/stretchr/testify/mock"
)

// MockClient is an autogenerated mock type for the Client type
type MockClient struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *MockClient) Close() {
	_m.Called()
}

// DeleteNode provides a mock function with given fields: ctx, in, opts
func (_m *MockClient) DeleteNode(ctx context.Context, in *Node, opts ...grpc.CallOption) (*DeleteNodeResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *DeleteNodeResponse
	if rf, ok := ret.Get(0).(func(context.Context, *Node, ...grpc.CallOption) *DeleteNodeResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*DeleteNodeResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *Node, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetNode provides a mock function with given fields: ctx, in, opts
func (_m *MockClient) GetNode(ctx context.Context, in *GetNodeRequest, opts ...grpc.CallOption) (*Node, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *Node
	if rf, ok := ret.Get(0).(func(context.Context, *GetNodeRequest, ...grpc.CallOption) *Node); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*Node)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *GetNodeRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListNodes provides a mock function with given fields: ctx, in, opts
func (_m *MockClient) ListNodes(ctx context.Context, in *ListNodesRequest, opts ...grpc.CallOption) (*ListNodesResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *ListNodesResponse
	if rf, ok := ret.Get(0).(func(context.Context, *ListNodesRequest, ...grpc.CallOption) *ListNodesResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*ListNodesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *ListNodesRequest, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PutNode provides a mock function with given fields: ctx, in, opts
func (_m *MockClient) PutNode(ctx context.Context, in *Node, opts ...grpc.CallOption) (*PutNodeResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *PutNodeResponse
	if rf, ok := ret.Get(0).(func(context.Context, *Node, ...grpc.CallOption) *PutNodeResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*PutNodeResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *Node, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WriteEvent provides a mock function with given fields: ctx, in, opts
func (_m *MockClient) WriteEvent(ctx context.Context, in *events.Event, opts ...grpc.CallOption) (*events.WriteEventResponse, error) {
	_va := make([]interface{}, len(opts))
	for _i := range opts {
		_va[_i] = opts[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx, in)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 *events.WriteEventResponse
	if rf, ok := ret.Get(0).(func(context.Context, *events.Event, ...grpc.CallOption) *events.WriteEventResponse); ok {
		r0 = rf(ctx, in, opts...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*events.WriteEventResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *events.Event, ...grpc.CallOption) error); ok {
		r1 = rf(ctx, in, opts...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
