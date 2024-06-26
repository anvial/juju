// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/agent/engine (interfaces: MetricSink)
//
// Generated by this command:
//
//	mockgen -typed -package addons_test -destination engine_mock_test.go github.com/juju/juju/agent/engine MetricSink
//

// Package addons_test is a generated GoMock package.
package addons_test

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockMetricSink is a mock of MetricSink interface.
type MockMetricSink struct {
	ctrl     *gomock.Controller
	recorder *MockMetricSinkMockRecorder
}

// MockMetricSinkMockRecorder is the mock recorder for MockMetricSink.
type MockMetricSinkMockRecorder struct {
	mock *MockMetricSink
}

// NewMockMetricSink creates a new mock instance.
func NewMockMetricSink(ctrl *gomock.Controller) *MockMetricSink {
	mock := &MockMetricSink{ctrl: ctrl}
	mock.recorder = &MockMetricSinkMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricSink) EXPECT() *MockMetricSinkMockRecorder {
	return m.recorder
}

// RecordStart mocks base method.
func (m *MockMetricSink) RecordStart(arg0 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "RecordStart", arg0)
}

// RecordStart indicates an expected call of RecordStart.
func (mr *MockMetricSinkMockRecorder) RecordStart(arg0 any) *MockMetricSinkRecordStartCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RecordStart", reflect.TypeOf((*MockMetricSink)(nil).RecordStart), arg0)
	return &MockMetricSinkRecordStartCall{Call: call}
}

// MockMetricSinkRecordStartCall wrap *gomock.Call
type MockMetricSinkRecordStartCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockMetricSinkRecordStartCall) Return() *MockMetricSinkRecordStartCall {
	c.Call = c.Call.Return()
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockMetricSinkRecordStartCall) Do(f func(string)) *MockMetricSinkRecordStartCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockMetricSinkRecordStartCall) DoAndReturn(f func(string)) *MockMetricSinkRecordStartCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Unregister mocks base method.
func (m *MockMetricSink) Unregister() bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unregister")
	ret0, _ := ret[0].(bool)
	return ret0
}

// Unregister indicates an expected call of Unregister.
func (mr *MockMetricSinkMockRecorder) Unregister() *MockMetricSinkUnregisterCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unregister", reflect.TypeOf((*MockMetricSink)(nil).Unregister))
	return &MockMetricSinkUnregisterCall{Call: call}
}

// MockMetricSinkUnregisterCall wrap *gomock.Call
type MockMetricSinkUnregisterCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockMetricSinkUnregisterCall) Return(arg0 bool) *MockMetricSinkUnregisterCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockMetricSinkUnregisterCall) Do(f func() bool) *MockMetricSinkUnregisterCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockMetricSinkUnregisterCall) DoAndReturn(f func() bool) *MockMetricSinkUnregisterCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
