// Code generated by mockery v2.33.0. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// SegmentService is an autogenerated mock type for the SegmentService type
type SegmentService struct {
	mock.Mock
}

type SegmentService_Expecter struct {
	mock *mock.Mock
}

func (_m *SegmentService) EXPECT() *SegmentService_Expecter {
	return &SegmentService_Expecter{mock: &_m.Mock}
}

// DeleteUserFromSegment provides a mock function with given fields: ctx, userID, slugs
func (_m *SegmentService) DeleteUserFromSegment(ctx context.Context, userID int64, slugs []string) error {
	ret := _m.Called(ctx, userID, slugs)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int64, []string) error); ok {
		r0 = rf(ctx, userID, slugs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SegmentService_DeleteUserFromSegment_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'DeleteUserFromSegment'
type SegmentService_DeleteUserFromSegment_Call struct {
	*mock.Call
}

// DeleteUserFromSegment is a helper method to define mock.On call
//   - ctx context.Context
//   - userID int64
//   - slugs []string
func (_e *SegmentService_Expecter) DeleteUserFromSegment(ctx interface{}, userID interface{}, slugs interface{}) *SegmentService_DeleteUserFromSegment_Call {
	return &SegmentService_DeleteUserFromSegment_Call{Call: _e.mock.On("DeleteUserFromSegment", ctx, userID, slugs)}
}

func (_c *SegmentService_DeleteUserFromSegment_Call) Run(run func(ctx context.Context, userID int64, slugs []string)) *SegmentService_DeleteUserFromSegment_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(int64), args[2].([]string))
	})
	return _c
}

func (_c *SegmentService_DeleteUserFromSegment_Call) Return(_a0 error) *SegmentService_DeleteUserFromSegment_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *SegmentService_DeleteUserFromSegment_Call) RunAndReturn(run func(context.Context, int64, []string) error) *SegmentService_DeleteUserFromSegment_Call {
	_c.Call.Return(run)
	return _c
}

// NewSegmentService creates a new instance of SegmentService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSegmentService(t interface {
	mock.TestingT
	Cleanup(func())
}) *SegmentService {
	mock := &SegmentService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
