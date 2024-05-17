// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/api/logsender (interfaces: LogWriter)
//
// Generated by this command:
//
//	mockgen -typed -package mocks -destination mocks/logwriter_mock.go github.com/juju/juju/api/logsender LogWriter
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	params "github.com/juju/juju/rpc/params"
	gomock "go.uber.org/mock/gomock"
)

// MockLogWriter is a mock of LogWriter interface.
type MockLogWriter struct {
	ctrl     *gomock.Controller
	recorder *MockLogWriterMockRecorder
}

// MockLogWriterMockRecorder is the mock recorder for MockLogWriter.
type MockLogWriterMockRecorder struct {
	mock *MockLogWriter
}

// NewMockLogWriter creates a new mock instance.
func NewMockLogWriter(ctrl *gomock.Controller) *MockLogWriter {
	mock := &MockLogWriter{ctrl: ctrl}
	mock.recorder = &MockLogWriterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogWriter) EXPECT() *MockLogWriterMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockLogWriter) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockLogWriterMockRecorder) Close() *MockLogWriterCloseCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockLogWriter)(nil).Close))
	return &MockLogWriterCloseCall{Call: call}
}

// MockLogWriterCloseCall wrap *gomock.Call
type MockLogWriterCloseCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockLogWriterCloseCall) Return(arg0 error) *MockLogWriterCloseCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockLogWriterCloseCall) Do(f func() error) *MockLogWriterCloseCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockLogWriterCloseCall) DoAndReturn(f func() error) *MockLogWriterCloseCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// WriteLog mocks base method.
func (m *MockLogWriter) WriteLog(arg0 *params.LogRecord) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteLog", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteLog indicates an expected call of WriteLog.
func (mr *MockLogWriterMockRecorder) WriteLog(arg0 any) *MockLogWriterWriteLogCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteLog", reflect.TypeOf((*MockLogWriter)(nil).WriteLog), arg0)
	return &MockLogWriterWriteLogCall{Call: call}
}

// MockLogWriterWriteLogCall wrap *gomock.Call
type MockLogWriterWriteLogCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockLogWriterWriteLogCall) Return(arg0 error) *MockLogWriterWriteLogCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockLogWriterWriteLogCall) Do(f func(*params.LogRecord) error) *MockLogWriterWriteLogCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockLogWriterWriteLogCall) DoAndReturn(f func(*params.LogRecord) error) *MockLogWriterWriteLogCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}