// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/core/objectstore (interfaces: ObjectStoreMetadata)

// Package objectstore is a generated GoMock package.
package objectstore

import (
	context "context"
	reflect "reflect"

	objectstore "github.com/juju/juju/core/objectstore"
	watcher "github.com/juju/juju/core/watcher"
	gomock "go.uber.org/mock/gomock"
)

// MockObjectStoreMetadata is a mock of ObjectStoreMetadata interface.
type MockObjectStoreMetadata struct {
	ctrl     *gomock.Controller
	recorder *MockObjectStoreMetadataMockRecorder
}

// MockObjectStoreMetadataMockRecorder is the mock recorder for MockObjectStoreMetadata.
type MockObjectStoreMetadataMockRecorder struct {
	mock *MockObjectStoreMetadata
}

// NewMockObjectStoreMetadata creates a new mock instance.
func NewMockObjectStoreMetadata(ctrl *gomock.Controller) *MockObjectStoreMetadata {
	mock := &MockObjectStoreMetadata{ctrl: ctrl}
	mock.recorder = &MockObjectStoreMetadataMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockObjectStoreMetadata) EXPECT() *MockObjectStoreMetadataMockRecorder {
	return m.recorder
}

// GetMetadata mocks base method.
func (m *MockObjectStoreMetadata) GetMetadata(arg0 context.Context, arg1 string) (objectstore.Metadata, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMetadata", arg0, arg1)
	ret0, _ := ret[0].(objectstore.Metadata)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMetadata indicates an expected call of GetMetadata.
func (mr *MockObjectStoreMetadataMockRecorder) GetMetadata(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMetadata", reflect.TypeOf((*MockObjectStoreMetadata)(nil).GetMetadata), arg0, arg1)
}

// PutMetadata mocks base method.
func (m *MockObjectStoreMetadata) PutMetadata(arg0 context.Context, arg1 objectstore.Metadata) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PutMetadata", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// PutMetadata indicates an expected call of PutMetadata.
func (mr *MockObjectStoreMetadataMockRecorder) PutMetadata(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PutMetadata", reflect.TypeOf((*MockObjectStoreMetadata)(nil).PutMetadata), arg0, arg1)
}

// RemoveMetadata mocks base method.
func (m *MockObjectStoreMetadata) RemoveMetadata(arg0 context.Context, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RemoveMetadata", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// RemoveMetadata indicates an expected call of RemoveMetadata.
func (mr *MockObjectStoreMetadataMockRecorder) RemoveMetadata(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RemoveMetadata", reflect.TypeOf((*MockObjectStoreMetadata)(nil).RemoveMetadata), arg0, arg1)
}

// Watch mocks base method.
func (m *MockObjectStoreMetadata) Watch() (watcher.Watcher[[]string], error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Watch")
	ret0, _ := ret[0].(watcher.Watcher[[]string])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Watch indicates an expected call of Watch.
func (mr *MockObjectStoreMetadataMockRecorder) Watch() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Watch", reflect.TypeOf((*MockObjectStoreMetadata)(nil).Watch))
}