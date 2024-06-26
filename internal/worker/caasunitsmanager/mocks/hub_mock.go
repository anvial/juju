// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/internal/worker/caasunitsmanager (interfaces: Hub)
//
// Generated by this command:
//
//	mockgen -typed -package mocks -destination mocks/hub_mock.go github.com/juju/juju/internal/worker/caasunitsmanager Hub
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockHub is a mock of Hub interface.
type MockHub struct {
	ctrl     *gomock.Controller
	recorder *MockHubMockRecorder
}

// MockHubMockRecorder is the mock recorder for MockHub.
type MockHubMockRecorder struct {
	mock *MockHub
}

// NewMockHub creates a new mock instance.
func NewMockHub(ctrl *gomock.Controller) *MockHub {
	mock := &MockHub{ctrl: ctrl}
	mock.recorder = &MockHubMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockHub) EXPECT() *MockHubMockRecorder {
	return m.recorder
}

// Publish mocks base method.
func (m *MockHub) Publish(arg0 string, arg1 any) func() {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Publish", arg0, arg1)
	ret0, _ := ret[0].(func())
	return ret0
}

// Publish indicates an expected call of Publish.
func (mr *MockHubMockRecorder) Publish(arg0, arg1 any) *MockHubPublishCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockHub)(nil).Publish), arg0, arg1)
	return &MockHubPublishCall{Call: call}
}

// MockHubPublishCall wrap *gomock.Call
type MockHubPublishCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHubPublishCall) Return(arg0 func()) *MockHubPublishCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHubPublishCall) Do(f func(string, any) func()) *MockHubPublishCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHubPublishCall) DoAndReturn(f func(string, any) func()) *MockHubPublishCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Subscribe mocks base method.
func (m *MockHub) Subscribe(arg0 string, arg1 func(string, any)) func() {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Subscribe", arg0, arg1)
	ret0, _ := ret[0].(func())
	return ret0
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockHubMockRecorder) Subscribe(arg0, arg1 any) *MockHubSubscribeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockHub)(nil).Subscribe), arg0, arg1)
	return &MockHubSubscribeCall{Call: call}
}

// MockHubSubscribeCall wrap *gomock.Call
type MockHubSubscribeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockHubSubscribeCall) Return(arg0 func()) *MockHubSubscribeCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockHubSubscribeCall) Do(f func(string, func(string, any)) func()) *MockHubSubscribeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockHubSubscribeCall) DoAndReturn(f func(string, func(string, any)) func()) *MockHubSubscribeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
