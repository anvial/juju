// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/environs (interfaces: ConnectorInfo)
//
// Generated by this command:
//
//	mockgen -typed -package apiserver_test -destination registration_environs_mock_test.go github.com/juju/juju/environs ConnectorInfo
//

// Package apiserver_test is a generated GoMock package.
package apiserver_test

import (
	context "context"
	reflect "reflect"

	proxy "github.com/juju/juju/internal/proxy"
	gomock "go.uber.org/mock/gomock"
)

// MockConnectorInfo is a mock of ConnectorInfo interface.
type MockConnectorInfo struct {
	ctrl     *gomock.Controller
	recorder *MockConnectorInfoMockRecorder
}

// MockConnectorInfoMockRecorder is the mock recorder for MockConnectorInfo.
type MockConnectorInfoMockRecorder struct {
	mock *MockConnectorInfo
}

// NewMockConnectorInfo creates a new mock instance.
func NewMockConnectorInfo(ctrl *gomock.Controller) *MockConnectorInfo {
	mock := &MockConnectorInfo{ctrl: ctrl}
	mock.recorder = &MockConnectorInfoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockConnectorInfo) EXPECT() *MockConnectorInfoMockRecorder {
	return m.recorder
}

// ConnectionProxyInfo mocks base method.
func (m *MockConnectorInfo) ConnectionProxyInfo(arg0 context.Context) (proxy.Proxier, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConnectionProxyInfo", arg0)
	ret0, _ := ret[0].(proxy.Proxier)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConnectionProxyInfo indicates an expected call of ConnectionProxyInfo.
func (mr *MockConnectorInfoMockRecorder) ConnectionProxyInfo(arg0 any) *MockConnectorInfoConnectionProxyInfoCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConnectionProxyInfo", reflect.TypeOf((*MockConnectorInfo)(nil).ConnectionProxyInfo), arg0)
	return &MockConnectorInfoConnectionProxyInfoCall{Call: call}
}

// MockConnectorInfoConnectionProxyInfoCall wrap *gomock.Call
type MockConnectorInfoConnectionProxyInfoCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockConnectorInfoConnectionProxyInfoCall) Return(arg0 proxy.Proxier, arg1 error) *MockConnectorInfoConnectionProxyInfoCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockConnectorInfoConnectionProxyInfoCall) Do(f func(context.Context) (proxy.Proxier, error)) *MockConnectorInfoConnectionProxyInfoCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockConnectorInfoConnectionProxyInfoCall) DoAndReturn(f func(context.Context) (proxy.Proxier, error)) *MockConnectorInfoConnectionProxyInfoCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
