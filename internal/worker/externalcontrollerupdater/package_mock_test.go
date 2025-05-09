// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/internal/worker/externalcontrollerupdater (interfaces: ExternalControllerWatcherClientCloser,ExternalControllerUpdaterClient)
//
// Generated by this command:
//
//	mockgen -package externalcontrollerupdater_test -destination package_mock_test.go github.com/juju/juju/internal/worker/externalcontrollerupdater ExternalControllerWatcherClientCloser,ExternalControllerUpdaterClient
//

// Package externalcontrollerupdater_test is a generated GoMock package.
package externalcontrollerupdater_test

import (
	context "context"
	reflect "reflect"

	crosscontroller "github.com/juju/juju/api/controller/crosscontroller"
	crossmodel "github.com/juju/juju/core/crossmodel"
	watcher "github.com/juju/juju/core/watcher"
	gomock "go.uber.org/mock/gomock"
)

// MockExternalControllerWatcherClientCloser is a mock of ExternalControllerWatcherClientCloser interface.
type MockExternalControllerWatcherClientCloser struct {
	ctrl     *gomock.Controller
	recorder *MockExternalControllerWatcherClientCloserMockRecorder
}

// MockExternalControllerWatcherClientCloserMockRecorder is the mock recorder for MockExternalControllerWatcherClientCloser.
type MockExternalControllerWatcherClientCloserMockRecorder struct {
	mock *MockExternalControllerWatcherClientCloser
}

// NewMockExternalControllerWatcherClientCloser creates a new mock instance.
func NewMockExternalControllerWatcherClientCloser(ctrl *gomock.Controller) *MockExternalControllerWatcherClientCloser {
	mock := &MockExternalControllerWatcherClientCloser{ctrl: ctrl}
	mock.recorder = &MockExternalControllerWatcherClientCloserMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExternalControllerWatcherClientCloser) EXPECT() *MockExternalControllerWatcherClientCloserMockRecorder {
	return m.recorder
}

// Close mocks base method.
func (m *MockExternalControllerWatcherClientCloser) Close() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Close")
	ret0, _ := ret[0].(error)
	return ret0
}

// Close indicates an expected call of Close.
func (mr *MockExternalControllerWatcherClientCloserMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockExternalControllerWatcherClientCloser)(nil).Close))
}

// ControllerInfo mocks base method.
func (m *MockExternalControllerWatcherClientCloser) ControllerInfo(arg0 context.Context) (*crosscontroller.ControllerInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ControllerInfo", arg0)
	ret0, _ := ret[0].(*crosscontroller.ControllerInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ControllerInfo indicates an expected call of ControllerInfo.
func (mr *MockExternalControllerWatcherClientCloserMockRecorder) ControllerInfo(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ControllerInfo", reflect.TypeOf((*MockExternalControllerWatcherClientCloser)(nil).ControllerInfo), arg0)
}

// WatchControllerInfo mocks base method.
func (m *MockExternalControllerWatcherClientCloser) WatchControllerInfo(arg0 context.Context) (watcher.Watcher[struct{}], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatchControllerInfo", arg0)
	ret0, _ := ret[0].(watcher.Watcher[struct{}])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WatchControllerInfo indicates an expected call of WatchControllerInfo.
func (mr *MockExternalControllerWatcherClientCloserMockRecorder) WatchControllerInfo(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchControllerInfo", reflect.TypeOf((*MockExternalControllerWatcherClientCloser)(nil).WatchControllerInfo), arg0)
}

// MockExternalControllerUpdaterClient is a mock of ExternalControllerUpdaterClient interface.
type MockExternalControllerUpdaterClient struct {
	ctrl     *gomock.Controller
	recorder *MockExternalControllerUpdaterClientMockRecorder
}

// MockExternalControllerUpdaterClientMockRecorder is the mock recorder for MockExternalControllerUpdaterClient.
type MockExternalControllerUpdaterClientMockRecorder struct {
	mock *MockExternalControllerUpdaterClient
}

// NewMockExternalControllerUpdaterClient creates a new mock instance.
func NewMockExternalControllerUpdaterClient(ctrl *gomock.Controller) *MockExternalControllerUpdaterClient {
	mock := &MockExternalControllerUpdaterClient{ctrl: ctrl}
	mock.recorder = &MockExternalControllerUpdaterClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockExternalControllerUpdaterClient) EXPECT() *MockExternalControllerUpdaterClientMockRecorder {
	return m.recorder
}

// ExternalControllerInfo mocks base method.
func (m *MockExternalControllerUpdaterClient) ExternalControllerInfo(arg0 context.Context, arg1 string) (*crossmodel.ControllerInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ExternalControllerInfo", arg0, arg1)
	ret0, _ := ret[0].(*crossmodel.ControllerInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ExternalControllerInfo indicates an expected call of ExternalControllerInfo.
func (mr *MockExternalControllerUpdaterClientMockRecorder) ExternalControllerInfo(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ExternalControllerInfo", reflect.TypeOf((*MockExternalControllerUpdaterClient)(nil).ExternalControllerInfo), arg0, arg1)
}

// SetExternalControllerInfo mocks base method.
func (m *MockExternalControllerUpdaterClient) SetExternalControllerInfo(arg0 context.Context, arg1 crossmodel.ControllerInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetExternalControllerInfo", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetExternalControllerInfo indicates an expected call of SetExternalControllerInfo.
func (mr *MockExternalControllerUpdaterClientMockRecorder) SetExternalControllerInfo(arg0, arg1 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetExternalControllerInfo", reflect.TypeOf((*MockExternalControllerUpdaterClient)(nil).SetExternalControllerInfo), arg0, arg1)
}

// WatchExternalControllers mocks base method.
func (m *MockExternalControllerUpdaterClient) WatchExternalControllers(arg0 context.Context) (watcher.Watcher[[]string], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WatchExternalControllers", arg0)
	ret0, _ := ret[0].(watcher.Watcher[[]string])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// WatchExternalControllers indicates an expected call of WatchExternalControllers.
func (mr *MockExternalControllerUpdaterClientMockRecorder) WatchExternalControllers(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WatchExternalControllers", reflect.TypeOf((*MockExternalControllerUpdaterClient)(nil).WatchExternalControllers), arg0)
}
