// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/internal/worker/controlleragentconfig (interfaces: ConfigWatcher)
//
// Generated by this command:
//
//	mockgen -typed -package dbaccessor -destination controllerconfig_mock_test.go github.com/juju/juju/internal/worker/controlleragentconfig ConfigWatcher
//

// Package dbaccessor is a generated GoMock package.
package dbaccessor

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockConfigWatcher is a mock of ConfigWatcher interface.
type MockConfigWatcher struct {
	ctrl     *gomock.Controller
	recorder *MockConfigWatcherMockRecorder
}

// MockConfigWatcherMockRecorder is the mock recorder for MockConfigWatcher.
type MockConfigWatcherMockRecorder struct {
	mock *MockConfigWatcher
}

// NewMockConfigWatcher creates a new mock instance.
func NewMockConfigWatcher(ctrl *gomock.Controller) *MockConfigWatcher {
	mock := &MockConfigWatcher{ctrl: ctrl}
	mock.recorder = &MockConfigWatcherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConfigWatcher) EXPECT() *MockConfigWatcherMockRecorder {
	return m.recorder
}

// Changes mocks base method.
func (m *MockConfigWatcher) Changes() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Changes")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Changes indicates an expected call of Changes.
func (mr *MockConfigWatcherMockRecorder) Changes() *MockConfigWatcherChangesCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Changes", reflect.TypeOf((*MockConfigWatcher)(nil).Changes))
	return &MockConfigWatcherChangesCall{Call: call}
}

// MockConfigWatcherChangesCall wrap *gomock.Call
type MockConfigWatcherChangesCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockConfigWatcherChangesCall) Return(arg0 <-chan struct{}) *MockConfigWatcherChangesCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockConfigWatcherChangesCall) Do(f func() <-chan struct{}) *MockConfigWatcherChangesCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockConfigWatcherChangesCall) DoAndReturn(f func() <-chan struct{}) *MockConfigWatcherChangesCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Done mocks base method.
func (m *MockConfigWatcher) Done() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Done")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Done indicates an expected call of Done.
func (mr *MockConfigWatcherMockRecorder) Done() *MockConfigWatcherDoneCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Done", reflect.TypeOf((*MockConfigWatcher)(nil).Done))
	return &MockConfigWatcherDoneCall{Call: call}
}

// MockConfigWatcherDoneCall wrap *gomock.Call
type MockConfigWatcherDoneCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockConfigWatcherDoneCall) Return(arg0 <-chan struct{}) *MockConfigWatcherDoneCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockConfigWatcherDoneCall) Do(f func() <-chan struct{}) *MockConfigWatcherDoneCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockConfigWatcherDoneCall) DoAndReturn(f func() <-chan struct{}) *MockConfigWatcherDoneCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Unsubscribe mocks base method.
func (m *MockConfigWatcher) Unsubscribe() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Unsubscribe")
}

// Unsubscribe indicates an expected call of Unsubscribe.
func (mr *MockConfigWatcherMockRecorder) Unsubscribe() *MockConfigWatcherUnsubscribeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unsubscribe", reflect.TypeOf((*MockConfigWatcher)(nil).Unsubscribe))
	return &MockConfigWatcherUnsubscribeCall{Call: call}
}

// MockConfigWatcherUnsubscribeCall wrap *gomock.Call
type MockConfigWatcherUnsubscribeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockConfigWatcherUnsubscribeCall) Return() *MockConfigWatcherUnsubscribeCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockConfigWatcherUnsubscribeCall) Do(f func()) *MockConfigWatcherUnsubscribeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockConfigWatcherUnsubscribeCall) DoAndReturn(f func()) *MockConfigWatcherUnsubscribeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
