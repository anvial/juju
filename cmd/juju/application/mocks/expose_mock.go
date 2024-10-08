// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/cmd/juju/application (interfaces: ApplicationExposeAPI)
//
// Generated by this command:
//
//	mockgen -typed -package mocks -destination mocks/expose_mock.go github.com/juju/juju/cmd/juju/application ApplicationExposeAPI
//

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	params "github.com/juju/juju/rpc/params"
	gomock "go.uber.org/mock/gomock"
)

// MockApplicationExposeAPI is a mock of ApplicationExposeAPI interface.
type MockApplicationExposeAPI struct {
	ctrl     *gomock.Controller
	recorder *MockApplicationExposeAPIMockRecorder
}

// MockApplicationExposeAPIMockRecorder is the mock recorder for MockApplicationExposeAPI.
type MockApplicationExposeAPIMockRecorder struct {
	mock *MockApplicationExposeAPI
}

// NewMockApplicationExposeAPI creates a new mock instance.
func NewMockApplicationExposeAPI(ctrl *gomock.Controller) *MockApplicationExposeAPI {
	mock := &MockApplicationExposeAPI{ctrl: ctrl}
	mock.recorder = &MockApplicationExposeAPIMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockApplicationExposeAPI) EXPECT() *MockApplicationExposeAPIMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockApplicationExposeAPI) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockApplicationExposeAPIMockRecorder) Close() *MockApplicationExposeAPICloseCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockApplicationExposeAPI)(nil).Close))
	return &MockApplicationExposeAPICloseCall{Call: call}
}

// MockApplicationExposeAPICloseCall wrap *gomock.Call
type MockApplicationExposeAPICloseCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockApplicationExposeAPICloseCall) Return(arg0 error) *MockApplicationExposeAPICloseCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockApplicationExposeAPICloseCall) Do(f func() error) *MockApplicationExposeAPICloseCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockApplicationExposeAPICloseCall) DoAndReturn(f func() error) *MockApplicationExposeAPICloseCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Expose mocks base method.
func (m *MockApplicationExposeAPI) Expose(arg0 context.Context, arg1 string, arg2 map[string]params.ExposedEndpoint) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Expose", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Expose indicates an expected call of Expose.
func (mr *MockApplicationExposeAPIMockRecorder) Expose(arg0, arg1, arg2 any) *MockApplicationExposeAPIExposeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Expose", reflect.TypeOf((*MockApplicationExposeAPI)(nil).Expose), arg0, arg1, arg2)
	return &MockApplicationExposeAPIExposeCall{Call: call}
}

// MockApplicationExposeAPIExposeCall wrap *gomock.Call
type MockApplicationExposeAPIExposeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockApplicationExposeAPIExposeCall) Return(arg0 error) *MockApplicationExposeAPIExposeCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockApplicationExposeAPIExposeCall) Do(f func(context.Context, string, map[string]params.ExposedEndpoint) error) *MockApplicationExposeAPIExposeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockApplicationExposeAPIExposeCall) DoAndReturn(f func(context.Context, string, map[string]params.ExposedEndpoint) error) *MockApplicationExposeAPIExposeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Unexpose mocks base method.
func (m *MockApplicationExposeAPI) Unexpose(arg0 context.Context, arg1 string, arg2 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unexpose", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Unexpose indicates an expected call of Unexpose.
func (mr *MockApplicationExposeAPIMockRecorder) Unexpose(arg0, arg1, arg2 any) *MockApplicationExposeAPIUnexposeCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unexpose", reflect.TypeOf((*MockApplicationExposeAPI)(nil).Unexpose), arg0, arg1, arg2)
	return &MockApplicationExposeAPIUnexposeCall{Call: call}
}

// MockApplicationExposeAPIUnexposeCall wrap *gomock.Call
type MockApplicationExposeAPIUnexposeCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockApplicationExposeAPIUnexposeCall) Return(arg0 error) *MockApplicationExposeAPIUnexposeCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockApplicationExposeAPIUnexposeCall) Do(f func(context.Context, string, []string) error) *MockApplicationExposeAPIUnexposeCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockApplicationExposeAPIUnexposeCall) DoAndReturn(f func(context.Context, string, []string) error) *MockApplicationExposeAPIUnexposeCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
