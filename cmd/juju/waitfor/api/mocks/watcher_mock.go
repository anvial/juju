// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/cmd/juju/waitfor/api (interfaces: WatchAllAPI,AllWatcher)
//
// Generated by this command:
//
//	mockgen -typed -package mocks -destination mocks/watcher_mock.go github.com/juju/juju/cmd/juju/waitfor/api WatchAllAPI,AllWatcher
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	api "github.com/juju/juju/cmd/juju/waitfor/api"
	params "github.com/juju/juju/rpc/params"
	gomock "go.uber.org/mock/gomock"
)

// MockWatchAllAPI is a mock of WatchAllAPI interface.
type MockWatchAllAPI struct {
	ctrl     *gomock.Controller
	recorder *MockWatchAllAPIMockRecorder
}

// MockWatchAllAPIMockRecorder is the mock recorder for MockWatchAllAPI.
type MockWatchAllAPIMockRecorder struct {
	mock *MockWatchAllAPI
}

// NewMockWatchAllAPI creates a new mock instance.
func NewMockWatchAllAPI(ctrl *gomock.Controller) *MockWatchAllAPI {
	mock := &MockWatchAllAPI{ctrl: ctrl}
	mock.recorder = &MockWatchAllAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWatchAllAPI) EXPECT() *MockWatchAllAPIMockRecorder {
	return m.recorder
}

// WatchAll mocks base method.
func (m *MockWatchAllAPI) WatchAll() (api.AllWatcher, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatchAll")
	ret0, _ := ret[0].(api.AllWatcher)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WatchAll indicates an expected call of WatchAll.
func (mr *MockWatchAllAPIMockRecorder) WatchAll() *MockWatchAllAPIWatchAllCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchAll", reflect.TypeOf((*MockWatchAllAPI)(nil).WatchAll))
	return &MockWatchAllAPIWatchAllCall{Call: call}
}

// MockWatchAllAPIWatchAllCall wrap *gomock.Call
type MockWatchAllAPIWatchAllCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockWatchAllAPIWatchAllCall) Return(arg0 api.AllWatcher, arg1 error) *MockWatchAllAPIWatchAllCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockWatchAllAPIWatchAllCall) Do(f func() (api.AllWatcher, error)) *MockWatchAllAPIWatchAllCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockWatchAllAPIWatchAllCall) DoAndReturn(f func() (api.AllWatcher, error)) *MockWatchAllAPIWatchAllCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MockAllWatcher is a mock of AllWatcher interface.
type MockAllWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockAllWatcherMockRecorder
}

// MockAllWatcherMockRecorder is the mock recorder for MockAllWatcher.
type MockAllWatcherMockRecorder struct {
	mock *MockAllWatcher
}

// NewMockAllWatcher creates a new mock instance.
func NewMockAllWatcher(ctrl *gomock.Controller) *MockAllWatcher {
	mock := &MockAllWatcher{ctrl: ctrl}
	mock.recorder = &MockAllWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAllWatcher) EXPECT() *MockAllWatcherMockRecorder {
	return m.recorder
}

// Next mocks base method.
func (m *MockAllWatcher) Next() ([]params.Delta, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Next")
	ret0, _ := ret[0].([]params.Delta)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Next indicates an expected call of Next.
func (mr *MockAllWatcherMockRecorder) Next() *MockAllWatcherNextCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Next", reflect.TypeOf((*MockAllWatcher)(nil).Next))
	return &MockAllWatcherNextCall{Call: call}
}

// MockAllWatcherNextCall wrap *gomock.Call
type MockAllWatcherNextCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockAllWatcherNextCall) Return(arg0 []params.Delta, arg1 error) *MockAllWatcherNextCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockAllWatcherNextCall) Do(f func() ([]params.Delta, error)) *MockAllWatcherNextCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockAllWatcherNextCall) DoAndReturn(f func() ([]params.Delta, error)) *MockAllWatcherNextCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Stop mocks base method.
func (m *MockAllWatcher) Stop() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Stop")
	ret0, _ := ret[0].(error)
	return ret0
}

// Stop indicates an expected call of Stop.
func (mr *MockAllWatcherMockRecorder) Stop() *MockAllWatcherStopCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stop", reflect.TypeOf((*MockAllWatcher)(nil).Stop))
	return &MockAllWatcherStopCall{Call: call}
}

// MockAllWatcherStopCall wrap *gomock.Call
type MockAllWatcherStopCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockAllWatcherStopCall) Return(arg0 error) *MockAllWatcherStopCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockAllWatcherStopCall) Do(f func() error) *MockAllWatcherStopCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockAllWatcherStopCall) DoAndReturn(f func() error) *MockAllWatcherStopCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
