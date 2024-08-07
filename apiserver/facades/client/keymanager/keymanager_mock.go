// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/apiserver/facades/client/keymanager (interfaces: BlockChecker)
//
// Generated by this command:
//
//	mockgen -typed -package keymanager -destination keymanager_mock.go github.com/juju/juju/apiserver/facades/client/keymanager BlockChecker
//

// Package keymanager is a generated GoMock package.
package keymanager

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockBlockChecker is a mock of BlockChecker interface.
type MockBlockChecker struct {
	ctrl     *gomock.Controller
	recorder *MockBlockCheckerMockRecorder
}

// MockBlockCheckerMockRecorder is the mock recorder for MockBlockChecker.
type MockBlockCheckerMockRecorder struct {
	mock *MockBlockChecker
}

// NewMockBlockChecker creates a new mock instance.
func NewMockBlockChecker(ctrl *gomock.Controller) *MockBlockChecker {
	mock := &MockBlockChecker{ctrl: ctrl}
	mock.recorder = &MockBlockCheckerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockBlockChecker) EXPECT() *MockBlockCheckerMockRecorder {
	return m.recorder
}

// ChangeAllowed mocks base method.
func (m *MockBlockChecker) ChangeAllowed(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChangeAllowed", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ChangeAllowed indicates an expected call of ChangeAllowed.
func (mr *MockBlockCheckerMockRecorder) ChangeAllowed(arg0 any) *MockBlockCheckerChangeAllowedCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChangeAllowed", reflect.TypeOf((*MockBlockChecker)(nil).ChangeAllowed), arg0)
	return &MockBlockCheckerChangeAllowedCall{Call: call}
}

// MockBlockCheckerChangeAllowedCall wrap *gomock.Call
type MockBlockCheckerChangeAllowedCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBlockCheckerChangeAllowedCall) Return(arg0 error) *MockBlockCheckerChangeAllowedCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBlockCheckerChangeAllowedCall) Do(f func(context.Context) error) *MockBlockCheckerChangeAllowedCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBlockCheckerChangeAllowedCall) DoAndReturn(f func(context.Context) error) *MockBlockCheckerChangeAllowedCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// RemoveAllowed mocks base method.
func (m *MockBlockChecker) RemoveAllowed(arg0 context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveAllowed", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveAllowed indicates an expected call of RemoveAllowed.
func (mr *MockBlockCheckerMockRecorder) RemoveAllowed(arg0 any) *MockBlockCheckerRemoveAllowedCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveAllowed", reflect.TypeOf((*MockBlockChecker)(nil).RemoveAllowed), arg0)
	return &MockBlockCheckerRemoveAllowedCall{Call: call}
}

// MockBlockCheckerRemoveAllowedCall wrap *gomock.Call
type MockBlockCheckerRemoveAllowedCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockBlockCheckerRemoveAllowedCall) Return(arg0 error) *MockBlockCheckerRemoveAllowedCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockBlockCheckerRemoveAllowedCall) Do(f func(context.Context) error) *MockBlockCheckerRemoveAllowedCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockBlockCheckerRemoveAllowedCall) DoAndReturn(f func(context.Context) error) *MockBlockCheckerRemoveAllowedCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
