// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/core/changestream (interfaces: Subscription,WatchableDB,EventSource)
//
// Generated by this command:
//
//	mockgen -typed -package eventsource -destination changestream_mock_test.go github.com/juju/juju/core/changestream Subscription,WatchableDB,EventSource
//

// Package eventsource is a generated GoMock package.
package eventsource

import (
	context "context"
	sql "database/sql"
	reflect "reflect"

	sqlair "github.com/canonical/sqlair"
	changestream "github.com/juju/juju/core/changestream"
	gomock "go.uber.org/mock/gomock"
)

// MockSubscription is a mock of Subscription interface.
type MockSubscription struct {
	ctrl     *gomock.Controller
	recorder *MockSubscriptionMockRecorder
}

// MockSubscriptionMockRecorder is the mock recorder for MockSubscription.
type MockSubscriptionMockRecorder struct {
	mock *MockSubscription
}

// NewMockSubscription creates a new mock instance.
func NewMockSubscription(ctrl *gomock.Controller) *MockSubscription {
	mock := &MockSubscription{ctrl: ctrl}
	mock.recorder = &MockSubscriptionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSubscription) EXPECT() *MockSubscriptionMockRecorder {
	return m.recorder
}

// Changes mocks base method.
func (m *MockSubscription) Changes() <-chan []changestream.ChangeEvent {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Changes")
	ret0, _ := ret[0].(<-chan []changestream.ChangeEvent)
	return ret0
}

// Changes indicates an expected call of Changes.
func (mr *MockSubscriptionMockRecorder) Changes() *MockSubscriptionChangesCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Changes", reflect.TypeOf((*MockSubscription)(nil).Changes))
	return &MockSubscriptionChangesCall{Call: call}
}

// MockSubscriptionChangesCall wrap *gomock.Call
type MockSubscriptionChangesCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockSubscriptionChangesCall) Return(arg0 <-chan []changestream.ChangeEvent) *MockSubscriptionChangesCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockSubscriptionChangesCall) Do(f func() <-chan []changestream.ChangeEvent) *MockSubscriptionChangesCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockSubscriptionChangesCall) DoAndReturn(f func() <-chan []changestream.ChangeEvent) *MockSubscriptionChangesCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Done mocks base method.
func (m *MockSubscription) Done() <-chan struct{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Done")
	ret0, _ := ret[0].(<-chan struct{})
	return ret0
}

// Done indicates an expected call of Done.
func (mr *MockSubscriptionMockRecorder) Done() *MockSubscriptionDoneCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Done", reflect.TypeOf((*MockSubscription)(nil).Done))
	return &MockSubscriptionDoneCall{Call: call}
}

// MockSubscriptionDoneCall wrap *gomock.Call
type MockSubscriptionDoneCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockSubscriptionDoneCall) Return(arg0 <-chan struct{}) *MockSubscriptionDoneCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockSubscriptionDoneCall) Do(f func() <-chan struct{}) *MockSubscriptionDoneCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockSubscriptionDoneCall) DoAndReturn(f func() <-chan struct{}) *MockSubscriptionDoneCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Kill mocks base method.
func (m *MockSubscription) Kill() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Kill")
}

// Kill indicates an expected call of Kill.
func (mr *MockSubscriptionMockRecorder) Kill() *MockSubscriptionKillCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Kill", reflect.TypeOf((*MockSubscription)(nil).Kill))
	return &MockSubscriptionKillCall{Call: call}
}

// MockSubscriptionKillCall wrap *gomock.Call
type MockSubscriptionKillCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockSubscriptionKillCall) Return() *MockSubscriptionKillCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockSubscriptionKillCall) Do(f func()) *MockSubscriptionKillCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockSubscriptionKillCall) DoAndReturn(f func()) *MockSubscriptionKillCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Wait mocks base method.
func (m *MockSubscription) Wait() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Wait")
	ret0, _ := ret[0].(error)
	return ret0
}

// Wait indicates an expected call of Wait.
func (mr *MockSubscriptionMockRecorder) Wait() *MockSubscriptionWaitCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockSubscription)(nil).Wait))
	return &MockSubscriptionWaitCall{Call: call}
}

// MockSubscriptionWaitCall wrap *gomock.Call
type MockSubscriptionWaitCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockSubscriptionWaitCall) Return(arg0 error) *MockSubscriptionWaitCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockSubscriptionWaitCall) Do(f func() error) *MockSubscriptionWaitCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockSubscriptionWaitCall) DoAndReturn(f func() error) *MockSubscriptionWaitCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MockWatchableDB is a mock of WatchableDB interface.
type MockWatchableDB struct {
	ctrl     *gomock.Controller
	recorder *MockWatchableDBMockRecorder
}

// MockWatchableDBMockRecorder is the mock recorder for MockWatchableDB.
type MockWatchableDBMockRecorder struct {
	mock *MockWatchableDB
}

// NewMockWatchableDB creates a new mock instance.
func NewMockWatchableDB(ctrl *gomock.Controller) *MockWatchableDB {
	mock := &MockWatchableDB{ctrl: ctrl}
	mock.recorder = &MockWatchableDBMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWatchableDB) EXPECT() *MockWatchableDBMockRecorder {
	return m.recorder
}

// StdTxn mocks base method.
func (m *MockWatchableDB) StdTxn(arg0 context.Context, arg1 func(context.Context, *sql.Tx) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "StdTxn", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// StdTxn indicates an expected call of StdTxn.
func (mr *MockWatchableDBMockRecorder) StdTxn(arg0, arg1 any) *MockWatchableDBStdTxnCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StdTxn", reflect.TypeOf((*MockWatchableDB)(nil).StdTxn), arg0, arg1)
	return &MockWatchableDBStdTxnCall{Call: call}
}

// MockWatchableDBStdTxnCall wrap *gomock.Call
type MockWatchableDBStdTxnCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockWatchableDBStdTxnCall) Return(arg0 error) *MockWatchableDBStdTxnCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockWatchableDBStdTxnCall) Do(f func(context.Context, func(context.Context, *sql.Tx) error) error) *MockWatchableDBStdTxnCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockWatchableDBStdTxnCall) DoAndReturn(f func(context.Context, func(context.Context, *sql.Tx) error) error) *MockWatchableDBStdTxnCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Subscribe mocks base method.
func (m *MockWatchableDB) Subscribe(arg0 ...changestream.SubscriptionOption) (changestream.Subscription, error) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Subscribe", varargs...)
	ret0, _ := ret[0].(changestream.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockWatchableDBMockRecorder) Subscribe(arg0 ...any) *MockWatchableDBSubscribeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockWatchableDB)(nil).Subscribe), arg0...)
	return &MockWatchableDBSubscribeCall{Call: call}
}

// MockWatchableDBSubscribeCall wrap *gomock.Call
type MockWatchableDBSubscribeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockWatchableDBSubscribeCall) Return(arg0 changestream.Subscription, arg1 error) *MockWatchableDBSubscribeCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockWatchableDBSubscribeCall) Do(f func(...changestream.SubscriptionOption) (changestream.Subscription, error)) *MockWatchableDBSubscribeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockWatchableDBSubscribeCall) DoAndReturn(f func(...changestream.SubscriptionOption) (changestream.Subscription, error)) *MockWatchableDBSubscribeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Txn mocks base method.
func (m *MockWatchableDB) Txn(arg0 context.Context, arg1 func(context.Context, *sqlair.TX) error) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Txn", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// Txn indicates an expected call of Txn.
func (mr *MockWatchableDBMockRecorder) Txn(arg0, arg1 any) *MockWatchableDBTxnCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Txn", reflect.TypeOf((*MockWatchableDB)(nil).Txn), arg0, arg1)
	return &MockWatchableDBTxnCall{Call: call}
}

// MockWatchableDBTxnCall wrap *gomock.Call
type MockWatchableDBTxnCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockWatchableDBTxnCall) Return(arg0 error) *MockWatchableDBTxnCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockWatchableDBTxnCall) Do(f func(context.Context, func(context.Context, *sqlair.TX) error) error) *MockWatchableDBTxnCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockWatchableDBTxnCall) DoAndReturn(f func(context.Context, func(context.Context, *sqlair.TX) error) error) *MockWatchableDBTxnCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MockEventSource is a mock of EventSource interface.
type MockEventSource struct {
	ctrl     *gomock.Controller
	recorder *MockEventSourceMockRecorder
}

// MockEventSourceMockRecorder is the mock recorder for MockEventSource.
type MockEventSourceMockRecorder struct {
	mock *MockEventSource
}

// NewMockEventSource creates a new mock instance.
func NewMockEventSource(ctrl *gomock.Controller) *MockEventSource {
	mock := &MockEventSource{ctrl: ctrl}
	mock.recorder = &MockEventSourceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEventSource) EXPECT() *MockEventSourceMockRecorder {
	return m.recorder
}

// Subscribe mocks base method.
func (m *MockEventSource) Subscribe(arg0 ...changestream.SubscriptionOption) (changestream.Subscription, error) {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Subscribe", varargs...)
	ret0, _ := ret[0].(changestream.Subscription)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Subscribe indicates an expected call of Subscribe.
func (mr *MockEventSourceMockRecorder) Subscribe(arg0 ...any) *MockEventSourceSubscribeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Subscribe", reflect.TypeOf((*MockEventSource)(nil).Subscribe), arg0...)
	return &MockEventSourceSubscribeCall{Call: call}
}

// MockEventSourceSubscribeCall wrap *gomock.Call
type MockEventSourceSubscribeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockEventSourceSubscribeCall) Return(arg0 changestream.Subscription, arg1 error) *MockEventSourceSubscribeCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockEventSourceSubscribeCall) Do(f func(...changestream.SubscriptionOption) (changestream.Subscription, error)) *MockEventSourceSubscribeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockEventSourceSubscribeCall) DoAndReturn(f func(...changestream.SubscriptionOption) (changestream.Subscription, error)) *MockEventSourceSubscribeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
