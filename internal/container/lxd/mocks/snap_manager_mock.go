// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/internal/container/lxd (interfaces: SnapManager)
//
// Generated by this command:
//
//	mockgen -typed -package mocks -destination mocks/snap_manager_mock.go github.com/juju/juju/internal/container/lxd SnapManager
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockSnapManager is a mock of SnapManager interface.
type MockSnapManager struct {
	ctrl     *gomock.Controller
	recorder *MockSnapManagerMockRecorder
}

// MockSnapManagerMockRecorder is the mock recorder for MockSnapManager.
type MockSnapManagerMockRecorder struct {
	mock *MockSnapManager
}

// NewMockSnapManager creates a new mock instance.
func NewMockSnapManager(ctrl *gomock.Controller) *MockSnapManager {
	mock := &MockSnapManager{ctrl: ctrl}
	mock.recorder = &MockSnapManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockSnapManager) EXPECT() *MockSnapManagerMockRecorder {
	return m.recorder
}

// ChangeChannel mocks base method.
func (m *MockSnapManager) ChangeChannel(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChangeChannel", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ChangeChannel indicates an expected call of ChangeChannel.
func (mr *MockSnapManagerMockRecorder) ChangeChannel(arg0, arg1 any) *MockSnapManagerChangeChannelCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChangeChannel", reflect.TypeOf((*MockSnapManager)(nil).ChangeChannel), arg0, arg1)
	return &MockSnapManagerChangeChannelCall{Call: call}
}

// MockSnapManagerChangeChannelCall wrap *gomock.Call
type MockSnapManagerChangeChannelCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockSnapManagerChangeChannelCall) Return(arg0 error) *MockSnapManagerChangeChannelCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockSnapManagerChangeChannelCall) Do(f func(string, string) error) *MockSnapManagerChangeChannelCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockSnapManagerChangeChannelCall) DoAndReturn(f func(string, string) error) *MockSnapManagerChangeChannelCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// InstalledChannel mocks base method.
func (m *MockSnapManager) InstalledChannel(arg0 string) string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InstalledChannel", arg0)
	ret0, _ := ret[0].(string)
	return ret0
}

// InstalledChannel indicates an expected call of InstalledChannel.
func (mr *MockSnapManagerMockRecorder) InstalledChannel(arg0 any) *MockSnapManagerInstalledChannelCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InstalledChannel", reflect.TypeOf((*MockSnapManager)(nil).InstalledChannel), arg0)
	return &MockSnapManagerInstalledChannelCall{Call: call}
}

// MockSnapManagerInstalledChannelCall wrap *gomock.Call
type MockSnapManagerInstalledChannelCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockSnapManagerInstalledChannelCall) Return(arg0 string) *MockSnapManagerInstalledChannelCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockSnapManagerInstalledChannelCall) Do(f func(string) string) *MockSnapManagerInstalledChannelCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockSnapManagerInstalledChannelCall) DoAndReturn(f func(string) string) *MockSnapManagerInstalledChannelCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
