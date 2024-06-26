// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/apiserver/logsink (interfaces: MetricsCollector)
//
// Generated by this command:
//
//	mockgen -typed -package mocks -destination mocks/metrics_collector_mock.go github.com/juju/juju/apiserver/logsink MetricsCollector
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	prometheus "github.com/prometheus/client_golang/prometheus"
	gomock "go.uber.org/mock/gomock"
)

// MockMetricsCollector is a mock of MetricsCollector interface.
type MockMetricsCollector struct {
	ctrl     *gomock.Controller
	recorder *MockMetricsCollectorMockRecorder
}

// MockMetricsCollectorMockRecorder is the mock recorder for MockMetricsCollector.
type MockMetricsCollectorMockRecorder struct {
	mock *MockMetricsCollector
}

// NewMockMetricsCollector creates a new mock instance.
func NewMockMetricsCollector(ctrl *gomock.Controller) *MockMetricsCollector {
	mock := &MockMetricsCollector{ctrl: ctrl}
	mock.recorder = &MockMetricsCollectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricsCollector) EXPECT() *MockMetricsCollectorMockRecorder {
	return m.recorder
}

// Connections mocks base method.
func (m *MockMetricsCollector) Connections() prometheus.Gauge {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Connections")
	ret0, _ := ret[0].(prometheus.Gauge)
	return ret0
}

// Connections indicates an expected call of Connections.
func (mr *MockMetricsCollectorMockRecorder) Connections() *MockMetricsCollectorConnectionsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Connections", reflect.TypeOf((*MockMetricsCollector)(nil).Connections))
	return &MockMetricsCollectorConnectionsCall{Call: call}
}

// MockMetricsCollectorConnectionsCall wrap *gomock.Call
type MockMetricsCollectorConnectionsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockMetricsCollectorConnectionsCall) Return(arg0 prometheus.Gauge) *MockMetricsCollectorConnectionsCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockMetricsCollectorConnectionsCall) Do(f func() prometheus.Gauge) *MockMetricsCollectorConnectionsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockMetricsCollectorConnectionsCall) DoAndReturn(f func() prometheus.Gauge) *MockMetricsCollectorConnectionsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// LogReadCount mocks base method.
func (m *MockMetricsCollector) LogReadCount(arg0, arg1 string) prometheus.Counter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LogReadCount", arg0, arg1)
	ret0, _ := ret[0].(prometheus.Counter)
	return ret0
}

// LogReadCount indicates an expected call of LogReadCount.
func (mr *MockMetricsCollectorMockRecorder) LogReadCount(arg0, arg1 any) *MockMetricsCollectorLogReadCountCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogReadCount", reflect.TypeOf((*MockMetricsCollector)(nil).LogReadCount), arg0, arg1)
	return &MockMetricsCollectorLogReadCountCall{Call: call}
}

// MockMetricsCollectorLogReadCountCall wrap *gomock.Call
type MockMetricsCollectorLogReadCountCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockMetricsCollectorLogReadCountCall) Return(arg0 prometheus.Counter) *MockMetricsCollectorLogReadCountCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockMetricsCollectorLogReadCountCall) Do(f func(string, string) prometheus.Counter) *MockMetricsCollectorLogReadCountCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockMetricsCollectorLogReadCountCall) DoAndReturn(f func(string, string) prometheus.Counter) *MockMetricsCollectorLogReadCountCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// LogWriteCount mocks base method.
func (m *MockMetricsCollector) LogWriteCount(arg0, arg1 string) prometheus.Counter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LogWriteCount", arg0, arg1)
	ret0, _ := ret[0].(prometheus.Counter)
	return ret0
}

// LogWriteCount indicates an expected call of LogWriteCount.
func (mr *MockMetricsCollectorMockRecorder) LogWriteCount(arg0, arg1 any) *MockMetricsCollectorLogWriteCountCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogWriteCount", reflect.TypeOf((*MockMetricsCollector)(nil).LogWriteCount), arg0, arg1)
	return &MockMetricsCollectorLogWriteCountCall{Call: call}
}

// MockMetricsCollectorLogWriteCountCall wrap *gomock.Call
type MockMetricsCollectorLogWriteCountCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockMetricsCollectorLogWriteCountCall) Return(arg0 prometheus.Counter) *MockMetricsCollectorLogWriteCountCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockMetricsCollectorLogWriteCountCall) Do(f func(string, string) prometheus.Counter) *MockMetricsCollectorLogWriteCountCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockMetricsCollectorLogWriteCountCall) DoAndReturn(f func(string, string) prometheus.Counter) *MockMetricsCollectorLogWriteCountCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// PingFailureCount mocks base method.
func (m *MockMetricsCollector) PingFailureCount(arg0 string) prometheus.Counter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PingFailureCount", arg0)
	ret0, _ := ret[0].(prometheus.Counter)
	return ret0
}

// PingFailureCount indicates an expected call of PingFailureCount.
func (mr *MockMetricsCollectorMockRecorder) PingFailureCount(arg0 any) *MockMetricsCollectorPingFailureCountCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PingFailureCount", reflect.TypeOf((*MockMetricsCollector)(nil).PingFailureCount), arg0)
	return &MockMetricsCollectorPingFailureCountCall{Call: call}
}

// MockMetricsCollectorPingFailureCountCall wrap *gomock.Call
type MockMetricsCollectorPingFailureCountCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockMetricsCollectorPingFailureCountCall) Return(arg0 prometheus.Counter) *MockMetricsCollectorPingFailureCountCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockMetricsCollectorPingFailureCountCall) Do(f func(string) prometheus.Counter) *MockMetricsCollectorPingFailureCountCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockMetricsCollectorPingFailureCountCall) DoAndReturn(f func(string) prometheus.Counter) *MockMetricsCollectorPingFailureCountCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// TotalConnections mocks base method.
func (m *MockMetricsCollector) TotalConnections() prometheus.Counter {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TotalConnections")
	ret0, _ := ret[0].(prometheus.Counter)
	return ret0
}

// TotalConnections indicates an expected call of TotalConnections.
func (mr *MockMetricsCollectorMockRecorder) TotalConnections() *MockMetricsCollectorTotalConnectionsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TotalConnections", reflect.TypeOf((*MockMetricsCollector)(nil).TotalConnections))
	return &MockMetricsCollectorTotalConnectionsCall{Call: call}
}

// MockMetricsCollectorTotalConnectionsCall wrap *gomock.Call
type MockMetricsCollectorTotalConnectionsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockMetricsCollectorTotalConnectionsCall) Return(arg0 prometheus.Counter) *MockMetricsCollectorTotalConnectionsCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockMetricsCollectorTotalConnectionsCall) Do(f func() prometheus.Counter) *MockMetricsCollectorTotalConnectionsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockMetricsCollectorTotalConnectionsCall) DoAndReturn(f func() prometheus.Counter) *MockMetricsCollectorTotalConnectionsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
