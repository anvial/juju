// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/apiserver/facades/client/charms/interfaces (interfaces: Downloader)
//
// Generated by this command:
//
//	mockgen -typed -package bootstrap -destination downloader_mock_test.go github.com/juju/juju/apiserver/facades/client/charms/interfaces Downloader
//

// Package bootstrap is a generated GoMock package.
package bootstrap

import (
	context "context"
	reflect "reflect"

	charm "github.com/juju/juju/core/charm"
	charm0 "github.com/juju/juju/internal/charm"
	gomock "go.uber.org/mock/gomock"
)

// MockDownloader is a mock of Downloader interface.
type MockDownloader struct {
	ctrl     *gomock.Controller
	recorder *MockDownloaderMockRecorder
}

// MockDownloaderMockRecorder is the mock recorder for MockDownloader.
type MockDownloaderMockRecorder struct {
	mock *MockDownloader
}

// NewMockDownloader creates a new mock instance.
func NewMockDownloader(ctrl *gomock.Controller) *MockDownloader {
	mock := &MockDownloader{ctrl: ctrl}
	mock.recorder = &MockDownloaderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDownloader) EXPECT() *MockDownloaderMockRecorder {
	return m.recorder
}

// DownloadAndStore mocks base method.
func (m *MockDownloader) DownloadAndStore(arg0 context.Context, arg1 *charm0.URL, arg2 charm.Origin, arg3 bool) (charm.Origin, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownloadAndStore", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].(charm.Origin)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownloadAndStore indicates an expected call of DownloadAndStore.
func (mr *MockDownloaderMockRecorder) DownloadAndStore(arg0, arg1, arg2, arg3 any) *MockDownloaderDownloadAndStoreCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownloadAndStore", reflect.TypeOf((*MockDownloader)(nil).DownloadAndStore), arg0, arg1, arg2, arg3)
	return &MockDownloaderDownloadAndStoreCall{Call: call}
}

// MockDownloaderDownloadAndStoreCall wrap *gomock.Call
type MockDownloaderDownloadAndStoreCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockDownloaderDownloadAndStoreCall) Return(arg0 charm.Origin, arg1 error) *MockDownloaderDownloadAndStoreCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockDownloaderDownloadAndStoreCall) Do(f func(context.Context, *charm0.URL, charm.Origin, bool) (charm.Origin, error)) *MockDownloaderDownloadAndStoreCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockDownloaderDownloadAndStoreCall) DoAndReturn(f func(context.Context, *charm0.URL, charm.Origin, bool) (charm.Origin, error)) *MockDownloaderDownloadAndStoreCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
