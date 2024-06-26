// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/state/watcher (interfaces: BaseWatcher)
//
// Generated by this command:
//
//	mockgen -typed -package mocks -destination mocks/watcher_mock.go github.com/juju/juju/state/watcher BaseWatcher
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	watcher "github.com/juju/juju/state/watcher"
	gomock "go.uber.org/mock/gomock"
)

// MockBaseWatcher is a mock of BaseWatcher interface.
type MockBaseWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockBaseWatcherMockRecorder
}

// MockBaseWatcherMockRecorder is the mock recorder for MockBaseWatcher.
type MockBaseWatcherMockRecorder struct {
	mock *MockBaseWatcher
}

// NewMockBaseWatcher creates a new mock instance.
func NewMockBaseWatcher(ctrl *gomock.Controller) *MockBaseWatcher {
	mock := &MockBaseWatcher{ctrl: ctrl}
	mock.recorder = &MockBaseWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBaseWatcher) EXPECT() *MockBaseWatcherMockRecorder {
	return m.recorder
}

// Dead mocks base method.
func (m *MockBaseWatcher) Dead() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Dead")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Dead indicates an expected call of Dead.
func (mr *MockBaseWatcherMockRecorder) Dead() *MockBaseWatcherDeadCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Dead", reflect.TypeOf((*MockBaseWatcher)(nil).Dead))
	return &MockBaseWatcherDeadCall{Call: call}
}

// MockBaseWatcherDeadCall wrap *gomock.Call
type MockBaseWatcherDeadCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBaseWatcherDeadCall) Return(arg0 <-chan struct{}) *MockBaseWatcherDeadCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBaseWatcherDeadCall) Do(f func() <-chan struct{}) *MockBaseWatcherDeadCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBaseWatcherDeadCall) DoAndReturn(f func() <-chan struct{}) *MockBaseWatcherDeadCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Err mocks base method.
func (m *MockBaseWatcher) Err() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Err")
	ret0, _ := ret[0].(error)
	return ret0
}

// Err indicates an expected call of Err.
func (mr *MockBaseWatcherMockRecorder) Err() *MockBaseWatcherErrCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Err", reflect.TypeOf((*MockBaseWatcher)(nil).Err))
	return &MockBaseWatcherErrCall{Call: call}
}

// MockBaseWatcherErrCall wrap *gomock.Call
type MockBaseWatcherErrCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBaseWatcherErrCall) Return(arg0 error) *MockBaseWatcherErrCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBaseWatcherErrCall) Do(f func() error) *MockBaseWatcherErrCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBaseWatcherErrCall) DoAndReturn(f func() error) *MockBaseWatcherErrCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Kill mocks base method.
func (m *MockBaseWatcher) Kill() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Kill")
}

// Kill indicates an expected call of Kill.
func (mr *MockBaseWatcherMockRecorder) Kill() *MockBaseWatcherKillCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kill", reflect.TypeOf((*MockBaseWatcher)(nil).Kill))
	return &MockBaseWatcherKillCall{Call: call}
}

// MockBaseWatcherKillCall wrap *gomock.Call
type MockBaseWatcherKillCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBaseWatcherKillCall) Return() *MockBaseWatcherKillCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBaseWatcherKillCall) Do(f func()) *MockBaseWatcherKillCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBaseWatcherKillCall) DoAndReturn(f func()) *MockBaseWatcherKillCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Unwatch mocks base method.
func (m *MockBaseWatcher) Unwatch(arg0 string, arg1 any, arg2 chan<- watcher.Change) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Unwatch", arg0, arg1, arg2)
}

// Unwatch indicates an expected call of Unwatch.
func (mr *MockBaseWatcherMockRecorder) Unwatch(arg0, arg1, arg2 any) *MockBaseWatcherUnwatchCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unwatch", reflect.TypeOf((*MockBaseWatcher)(nil).Unwatch), arg0, arg1, arg2)
	return &MockBaseWatcherUnwatchCall{Call: call}
}

// MockBaseWatcherUnwatchCall wrap *gomock.Call
type MockBaseWatcherUnwatchCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBaseWatcherUnwatchCall) Return() *MockBaseWatcherUnwatchCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBaseWatcherUnwatchCall) Do(f func(string, any, chan<- watcher.Change)) *MockBaseWatcherUnwatchCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBaseWatcherUnwatchCall) DoAndReturn(f func(string, any, chan<- watcher.Change)) *MockBaseWatcherUnwatchCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// UnwatchCollection mocks base method.
func (m *MockBaseWatcher) UnwatchCollection(arg0 string, arg1 chan<- watcher.Change) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UnwatchCollection", arg0, arg1)
}

// UnwatchCollection indicates an expected call of UnwatchCollection.
func (mr *MockBaseWatcherMockRecorder) UnwatchCollection(arg0, arg1 any) *MockBaseWatcherUnwatchCollectionCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnwatchCollection", reflect.TypeOf((*MockBaseWatcher)(nil).UnwatchCollection), arg0, arg1)
	return &MockBaseWatcherUnwatchCollectionCall{Call: call}
}

// MockBaseWatcherUnwatchCollectionCall wrap *gomock.Call
type MockBaseWatcherUnwatchCollectionCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBaseWatcherUnwatchCollectionCall) Return() *MockBaseWatcherUnwatchCollectionCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBaseWatcherUnwatchCollectionCall) Do(f func(string, chan<- watcher.Change)) *MockBaseWatcherUnwatchCollectionCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBaseWatcherUnwatchCollectionCall) DoAndReturn(f func(string, chan<- watcher.Change)) *MockBaseWatcherUnwatchCollectionCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Wait mocks base method.
func (m *MockBaseWatcher) Wait() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Wait")
	ret0, _ := ret[0].(error)
	return ret0
}

// Wait indicates an expected call of Wait.
func (mr *MockBaseWatcherMockRecorder) Wait() *MockBaseWatcherWaitCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockBaseWatcher)(nil).Wait))
	return &MockBaseWatcherWaitCall{Call: call}
}

// MockBaseWatcherWaitCall wrap *gomock.Call
type MockBaseWatcherWaitCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBaseWatcherWaitCall) Return(arg0 error) *MockBaseWatcherWaitCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBaseWatcherWaitCall) Do(f func() error) *MockBaseWatcherWaitCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBaseWatcherWaitCall) DoAndReturn(f func() error) *MockBaseWatcherWaitCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Watch mocks base method.
func (m *MockBaseWatcher) Watch(arg0 string, arg1 any, arg2 chan<- watcher.Change) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Watch", arg0, arg1, arg2)
}

// Watch indicates an expected call of Watch.
func (mr *MockBaseWatcherMockRecorder) Watch(arg0, arg1, arg2 any) *MockBaseWatcherWatchCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockBaseWatcher)(nil).Watch), arg0, arg1, arg2)
	return &MockBaseWatcherWatchCall{Call: call}
}

// MockBaseWatcherWatchCall wrap *gomock.Call
type MockBaseWatcherWatchCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBaseWatcherWatchCall) Return() *MockBaseWatcherWatchCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBaseWatcherWatchCall) Do(f func(string, any, chan<- watcher.Change)) *MockBaseWatcherWatchCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBaseWatcherWatchCall) DoAndReturn(f func(string, any, chan<- watcher.Change)) *MockBaseWatcherWatchCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// WatchCollection mocks base method.
func (m *MockBaseWatcher) WatchCollection(arg0 string, arg1 chan<- watcher.Change) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WatchCollection", arg0, arg1)
}

// WatchCollection indicates an expected call of WatchCollection.
func (mr *MockBaseWatcherMockRecorder) WatchCollection(arg0, arg1 any) *MockBaseWatcherWatchCollectionCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchCollection", reflect.TypeOf((*MockBaseWatcher)(nil).WatchCollection), arg0, arg1)
	return &MockBaseWatcherWatchCollectionCall{Call: call}
}

// MockBaseWatcherWatchCollectionCall wrap *gomock.Call
type MockBaseWatcherWatchCollectionCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBaseWatcherWatchCollectionCall) Return() *MockBaseWatcherWatchCollectionCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBaseWatcherWatchCollectionCall) Do(f func(string, chan<- watcher.Change)) *MockBaseWatcherWatchCollectionCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBaseWatcherWatchCollectionCall) DoAndReturn(f func(string, chan<- watcher.Change)) *MockBaseWatcherWatchCollectionCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// WatchCollectionWithFilter mocks base method.
func (m *MockBaseWatcher) WatchCollectionWithFilter(arg0 string, arg1 chan<- watcher.Change, arg2 func(any) bool) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WatchCollectionWithFilter", arg0, arg1, arg2)
}

// WatchCollectionWithFilter indicates an expected call of WatchCollectionWithFilter.
func (mr *MockBaseWatcherMockRecorder) WatchCollectionWithFilter(arg0, arg1, arg2 any) *MockBaseWatcherWatchCollectionWithFilterCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchCollectionWithFilter", reflect.TypeOf((*MockBaseWatcher)(nil).WatchCollectionWithFilter), arg0, arg1, arg2)
	return &MockBaseWatcherWatchCollectionWithFilterCall{Call: call}
}

// MockBaseWatcherWatchCollectionWithFilterCall wrap *gomock.Call
type MockBaseWatcherWatchCollectionWithFilterCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBaseWatcherWatchCollectionWithFilterCall) Return() *MockBaseWatcherWatchCollectionWithFilterCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBaseWatcherWatchCollectionWithFilterCall) Do(f func(string, chan<- watcher.Change, func(any) bool)) *MockBaseWatcherWatchCollectionWithFilterCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBaseWatcherWatchCollectionWithFilterCall) DoAndReturn(f func(string, chan<- watcher.Change, func(any) bool)) *MockBaseWatcherWatchCollectionWithFilterCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// WatchMulti mocks base method.
func (m *MockBaseWatcher) WatchMulti(arg0 string, arg1 []any, arg2 chan<- watcher.Change) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatchMulti", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// WatchMulti indicates an expected call of WatchMulti.
func (mr *MockBaseWatcherMockRecorder) WatchMulti(arg0, arg1, arg2 any) *MockBaseWatcherWatchMultiCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchMulti", reflect.TypeOf((*MockBaseWatcher)(nil).WatchMulti), arg0, arg1, arg2)
	return &MockBaseWatcherWatchMultiCall{Call: call}
}

// MockBaseWatcherWatchMultiCall wrap *gomock.Call
type MockBaseWatcherWatchMultiCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBaseWatcherWatchMultiCall) Return(arg0 error) *MockBaseWatcherWatchMultiCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBaseWatcherWatchMultiCall) Do(f func(string, []any, chan<- watcher.Change) error) *MockBaseWatcherWatchMultiCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBaseWatcherWatchMultiCall) DoAndReturn(f func(string, []any, chan<- watcher.Change) error) *MockBaseWatcherWatchMultiCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
