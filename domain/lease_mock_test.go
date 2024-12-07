// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/core/lease (interfaces: Token,LeaseCheckerWaiter,ModelLeaseManagerGetter)
//
// Generated by this command:
//
//	mockgen -typed -package domain -destination lease_mock_test.go github.com/juju/juju/core/lease Token,LeaseCheckerWaiter,ModelLeaseManagerGetter
//

// Package domain is a generated GoMock package.
package domain

import (
	context "context"
	reflect "reflect"

	lease "github.com/juju/juju/core/lease"
	gomock "go.uber.org/mock/gomock"
)

// MockToken is a mock of Token interface.
type MockToken struct {
	ctrl     *gomock.Controller
	recorder *MockTokenMockRecorder
}

// MockTokenMockRecorder is the mock recorder for MockToken.
type MockTokenMockRecorder struct {
	mock *MockToken
}

// NewMockToken creates a new mock instance.
func NewMockToken(ctrl *gomock.Controller) *MockToken {
	mock := &MockToken{ctrl: ctrl}
	mock.recorder = &MockTokenMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockToken) EXPECT() *MockTokenMockRecorder {
	return m.recorder
}

// Check mocks base method.
func (m *MockToken) Check() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Check")
	ret0, _ := ret[0].(error)
	return ret0
}

// Check indicates an expected call of Check.
func (mr *MockTokenMockRecorder) Check() *MockTokenCheckCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Check", reflect.TypeOf((*MockToken)(nil).Check))
	return &MockTokenCheckCall{Call: call}
}

// MockTokenCheckCall wrap *gomock.Call
type MockTokenCheckCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockTokenCheckCall) Return(arg0 error) *MockTokenCheckCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockTokenCheckCall) Do(f func() error) *MockTokenCheckCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockTokenCheckCall) DoAndReturn(f func() error) *MockTokenCheckCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MockLeaseCheckerWaiter is a mock of LeaseCheckerWaiter interface.
type MockLeaseCheckerWaiter struct {
	ctrl     *gomock.Controller
	recorder *MockLeaseCheckerWaiterMockRecorder
}

// MockLeaseCheckerWaiterMockRecorder is the mock recorder for MockLeaseCheckerWaiter.
type MockLeaseCheckerWaiterMockRecorder struct {
	mock *MockLeaseCheckerWaiter
}

// NewMockLeaseCheckerWaiter creates a new mock instance.
func NewMockLeaseCheckerWaiter(ctrl *gomock.Controller) *MockLeaseCheckerWaiter {
	mock := &MockLeaseCheckerWaiter{ctrl: ctrl}
	mock.recorder = &MockLeaseCheckerWaiterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLeaseCheckerWaiter) EXPECT() *MockLeaseCheckerWaiterMockRecorder {
	return m.recorder
}

// Token mocks base method.
func (m *MockLeaseCheckerWaiter) Token(arg0, arg1 string) lease.Token {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Token", arg0, arg1)
	ret0, _ := ret[0].(lease.Token)
	return ret0
}

// Token indicates an expected call of Token.
func (mr *MockLeaseCheckerWaiterMockRecorder) Token(arg0, arg1 any) *MockLeaseCheckerWaiterTokenCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Token", reflect.TypeOf((*MockLeaseCheckerWaiter)(nil).Token), arg0, arg1)
	return &MockLeaseCheckerWaiterTokenCall{Call: call}
}

// MockLeaseCheckerWaiterTokenCall wrap *gomock.Call
type MockLeaseCheckerWaiterTokenCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockLeaseCheckerWaiterTokenCall) Return(arg0 lease.Token) *MockLeaseCheckerWaiterTokenCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockLeaseCheckerWaiterTokenCall) Do(f func(string, string) lease.Token) *MockLeaseCheckerWaiterTokenCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockLeaseCheckerWaiterTokenCall) DoAndReturn(f func(string, string) lease.Token) *MockLeaseCheckerWaiterTokenCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// WaitUntilExpired mocks base method.
func (m *MockLeaseCheckerWaiter) WaitUntilExpired(arg0 context.Context, arg1 string, arg2 chan<- struct{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitUntilExpired", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitUntilExpired indicates an expected call of WaitUntilExpired.
func (mr *MockLeaseCheckerWaiterMockRecorder) WaitUntilExpired(arg0, arg1, arg2 any) *MockLeaseCheckerWaiterWaitUntilExpiredCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitUntilExpired", reflect.TypeOf((*MockLeaseCheckerWaiter)(nil).WaitUntilExpired), arg0, arg1, arg2)
	return &MockLeaseCheckerWaiterWaitUntilExpiredCall{Call: call}
}

// MockLeaseCheckerWaiterWaitUntilExpiredCall wrap *gomock.Call
type MockLeaseCheckerWaiterWaitUntilExpiredCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockLeaseCheckerWaiterWaitUntilExpiredCall) Return(arg0 error) *MockLeaseCheckerWaiterWaitUntilExpiredCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockLeaseCheckerWaiterWaitUntilExpiredCall) Do(f func(context.Context, string, chan<- struct{}) error) *MockLeaseCheckerWaiterWaitUntilExpiredCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockLeaseCheckerWaiterWaitUntilExpiredCall) DoAndReturn(f func(context.Context, string, chan<- struct{}) error) *MockLeaseCheckerWaiterWaitUntilExpiredCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MockModelLeaseManagerGetter is a mock of ModelLeaseManagerGetter interface.
type MockModelLeaseManagerGetter struct {
	ctrl     *gomock.Controller
	recorder *MockModelLeaseManagerGetterMockRecorder
}

// MockModelLeaseManagerGetterMockRecorder is the mock recorder for MockModelLeaseManagerGetter.
type MockModelLeaseManagerGetterMockRecorder struct {
	mock *MockModelLeaseManagerGetter
}

// NewMockModelLeaseManagerGetter creates a new mock instance.
func NewMockModelLeaseManagerGetter(ctrl *gomock.Controller) *MockModelLeaseManagerGetter {
	mock := &MockModelLeaseManagerGetter{ctrl: ctrl}
	mock.recorder = &MockModelLeaseManagerGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockModelLeaseManagerGetter) EXPECT() *MockModelLeaseManagerGetterMockRecorder {
	return m.recorder
}

// GetLeaseManager mocks base method.
func (m *MockModelLeaseManagerGetter) GetLeaseManager() (lease.LeaseCheckerWaiter, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetLeaseManager")
	ret0, _ := ret[0].(lease.LeaseCheckerWaiter)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetLeaseManager indicates an expected call of GetLeaseManager.
func (mr *MockModelLeaseManagerGetterMockRecorder) GetLeaseManager() *MockModelLeaseManagerGetterGetLeaseManagerCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetLeaseManager", reflect.TypeOf((*MockModelLeaseManagerGetter)(nil).GetLeaseManager))
	return &MockModelLeaseManagerGetterGetLeaseManagerCall{Call: call}
}

// MockModelLeaseManagerGetterGetLeaseManagerCall wrap *gomock.Call
type MockModelLeaseManagerGetterGetLeaseManagerCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelLeaseManagerGetterGetLeaseManagerCall) Return(arg0 lease.LeaseCheckerWaiter, arg1 error) *MockModelLeaseManagerGetterGetLeaseManagerCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelLeaseManagerGetterGetLeaseManagerCall) Do(f func() (lease.LeaseCheckerWaiter, error)) *MockModelLeaseManagerGetterGetLeaseManagerCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelLeaseManagerGetterGetLeaseManagerCall) DoAndReturn(f func() (lease.LeaseCheckerWaiter, error)) *MockModelLeaseManagerGetterGetLeaseManagerCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
