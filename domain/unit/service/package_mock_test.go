// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/domain/unit/service (interfaces: State)
//
// Generated by this command:
//
//	mockgen -typed -package service -destination package_mock_test.go github.com/juju/juju/domain/unit/service State
//

// Package service is a generated GoMock package.
package service

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockState is a mock of State interface.
type MockState struct {
	ctrl     *gomock.Controller
	recorder *MockStateMockRecorder
}

// MockStateMockRecorder is the mock recorder for MockState.
type MockStateMockRecorder struct {
	mock *MockState
}

// NewMockState creates a new mock instance.
func NewMockState(ctrl *gomock.Controller) *MockState {
	mock := &MockState{ctrl: ctrl}
	mock.recorder = &MockStateMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockState) EXPECT() *MockStateMockRecorder {
	return m.recorder
}

// DeleteUnit mocks base method.
func (m *MockState) DeleteUnit(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUnit", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUnit indicates an expected call of DeleteUnit.
func (mr *MockStateMockRecorder) DeleteUnit(arg0, arg1 any) *MockStateDeleteUnitCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUnit", reflect.TypeOf((*MockState)(nil).DeleteUnit), arg0, arg1)
	return &MockStateDeleteUnitCall{Call: call}
}

// MockStateDeleteUnitCall wrap *gomock.Call
type MockStateDeleteUnitCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockStateDeleteUnitCall) Return(arg0 error) *MockStateDeleteUnitCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockStateDeleteUnitCall) Do(f func(context.Context, string) error) *MockStateDeleteUnitCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockStateDeleteUnitCall) DoAndReturn(f func(context.Context, string) error) *MockStateDeleteUnitCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
