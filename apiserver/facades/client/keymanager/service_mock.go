// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/apiserver/facades/client/keymanager (interfaces: KeyManagerService,UserService)
//
// Generated by this command:
//
//	mockgen -typed -package keymanager -destination service_mock.go github.com/juju/juju/apiserver/facades/client/keymanager KeyManagerService,UserService
//

// Package keymanager is a generated GoMock package.
package keymanager

import (
	context "context"
	url "net/url"
	reflect "reflect"

	ssh "github.com/juju/juju/core/ssh"
	user "github.com/juju/juju/core/user"
	gomock "go.uber.org/mock/gomock"
)

// MockKeyManagerService is a mock of KeyManagerService interface.
type MockKeyManagerService struct {
	ctrl     *gomock.Controller
	recorder *MockKeyManagerServiceMockRecorder
}

// MockKeyManagerServiceMockRecorder is the mock recorder for MockKeyManagerService.
type MockKeyManagerServiceMockRecorder struct {
	mock *MockKeyManagerService
}

// NewMockKeyManagerService creates a new mock instance.
func NewMockKeyManagerService(ctrl *gomock.Controller) *MockKeyManagerService {
	mock := &MockKeyManagerService{ctrl: ctrl}
	mock.recorder = &MockKeyManagerServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockKeyManagerService) EXPECT() *MockKeyManagerServiceMockRecorder {
	return m.recorder
}

// AddPublicKeysForUser mocks base method.
func (m *MockKeyManagerService) AddPublicKeysForUser(arg0 context.Context, arg1 user.UUID, arg2 ...string) error {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "AddPublicKeysForUser", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// AddPublicKeysForUser indicates an expected call of AddPublicKeysForUser.
func (mr *MockKeyManagerServiceMockRecorder) AddPublicKeysForUser(arg0, arg1 any, arg2 ...any) *MockKeyManagerServiceAddPublicKeysForUserCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPublicKeysForUser", reflect.TypeOf((*MockKeyManagerService)(nil).AddPublicKeysForUser), varargs...)
	return &MockKeyManagerServiceAddPublicKeysForUserCall{Call: call}
}

// MockKeyManagerServiceAddPublicKeysForUserCall wrap *gomock.Call
type MockKeyManagerServiceAddPublicKeysForUserCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockKeyManagerServiceAddPublicKeysForUserCall) Return(arg0 error) *MockKeyManagerServiceAddPublicKeysForUserCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockKeyManagerServiceAddPublicKeysForUserCall) Do(f func(context.Context, user.UUID, ...string) error) *MockKeyManagerServiceAddPublicKeysForUserCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockKeyManagerServiceAddPublicKeysForUserCall) DoAndReturn(f func(context.Context, user.UUID, ...string) error) *MockKeyManagerServiceAddPublicKeysForUserCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// DeleteKeysForUser mocks base method.
func (m *MockKeyManagerService) DeleteKeysForUser(arg0 context.Context, arg1 user.UUID, arg2 ...string) error {
	m.ctrl.T.Helper()
	varargs := []any{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteKeysForUser", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteKeysForUser indicates an expected call of DeleteKeysForUser.
func (mr *MockKeyManagerServiceMockRecorder) DeleteKeysForUser(arg0, arg1 any, arg2 ...any) *MockKeyManagerServiceDeleteKeysForUserCall {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{arg0, arg1}, arg2...)
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteKeysForUser", reflect.TypeOf((*MockKeyManagerService)(nil).DeleteKeysForUser), varargs...)
	return &MockKeyManagerServiceDeleteKeysForUserCall{Call: call}
}

// MockKeyManagerServiceDeleteKeysForUserCall wrap *gomock.Call
type MockKeyManagerServiceDeleteKeysForUserCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockKeyManagerServiceDeleteKeysForUserCall) Return(arg0 error) *MockKeyManagerServiceDeleteKeysForUserCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockKeyManagerServiceDeleteKeysForUserCall) Do(f func(context.Context, user.UUID, ...string) error) *MockKeyManagerServiceDeleteKeysForUserCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockKeyManagerServiceDeleteKeysForUserCall) DoAndReturn(f func(context.Context, user.UUID, ...string) error) *MockKeyManagerServiceDeleteKeysForUserCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ImportPublicKeysForUser mocks base method.
func (m *MockKeyManagerService) ImportPublicKeysForUser(arg0 context.Context, arg1 user.UUID, arg2 *url.URL) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ImportPublicKeysForUser", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// ImportPublicKeysForUser indicates an expected call of ImportPublicKeysForUser.
func (mr *MockKeyManagerServiceMockRecorder) ImportPublicKeysForUser(arg0, arg1, arg2 any) *MockKeyManagerServiceImportPublicKeysForUserCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ImportPublicKeysForUser", reflect.TypeOf((*MockKeyManagerService)(nil).ImportPublicKeysForUser), arg0, arg1, arg2)
	return &MockKeyManagerServiceImportPublicKeysForUserCall{Call: call}
}

// MockKeyManagerServiceImportPublicKeysForUserCall wrap *gomock.Call
type MockKeyManagerServiceImportPublicKeysForUserCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockKeyManagerServiceImportPublicKeysForUserCall) Return(arg0 error) *MockKeyManagerServiceImportPublicKeysForUserCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockKeyManagerServiceImportPublicKeysForUserCall) Do(f func(context.Context, user.UUID, *url.URL) error) *MockKeyManagerServiceImportPublicKeysForUserCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockKeyManagerServiceImportPublicKeysForUserCall) DoAndReturn(f func(context.Context, user.UUID, *url.URL) error) *MockKeyManagerServiceImportPublicKeysForUserCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// ListPublicKeysForUser mocks base method.
func (m *MockKeyManagerService) ListPublicKeysForUser(arg0 context.Context, arg1 user.UUID) ([]ssh.PublicKey, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPublicKeysForUser", arg0, arg1)
	ret0, _ := ret[0].([]ssh.PublicKey)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPublicKeysForUser indicates an expected call of ListPublicKeysForUser.
func (mr *MockKeyManagerServiceMockRecorder) ListPublicKeysForUser(arg0, arg1 any) *MockKeyManagerServiceListPublicKeysForUserCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPublicKeysForUser", reflect.TypeOf((*MockKeyManagerService)(nil).ListPublicKeysForUser), arg0, arg1)
	return &MockKeyManagerServiceListPublicKeysForUserCall{Call: call}
}

// MockKeyManagerServiceListPublicKeysForUserCall wrap *gomock.Call
type MockKeyManagerServiceListPublicKeysForUserCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockKeyManagerServiceListPublicKeysForUserCall) Return(arg0 []ssh.PublicKey, arg1 error) *MockKeyManagerServiceListPublicKeysForUserCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockKeyManagerServiceListPublicKeysForUserCall) Do(f func(context.Context, user.UUID) ([]ssh.PublicKey, error)) *MockKeyManagerServiceListPublicKeysForUserCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockKeyManagerServiceListPublicKeysForUserCall) DoAndReturn(f func(context.Context, user.UUID) ([]ssh.PublicKey, error)) *MockKeyManagerServiceListPublicKeysForUserCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// GetUserByName mocks base method.
func (m *MockUserService) GetUserByName(arg0 context.Context, arg1 user.Name) (user.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByName", arg0, arg1)
	ret0, _ := ret[0].(user.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByName indicates an expected call of GetUserByName.
func (mr *MockUserServiceMockRecorder) GetUserByName(arg0, arg1 any) *MockUserServiceGetUserByNameCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByName", reflect.TypeOf((*MockUserService)(nil).GetUserByName), arg0, arg1)
	return &MockUserServiceGetUserByNameCall{Call: call}
}

// MockUserServiceGetUserByNameCall wrap *gomock.Call
type MockUserServiceGetUserByNameCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockUserServiceGetUserByNameCall) Return(arg0 user.User, arg1 error) *MockUserServiceGetUserByNameCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockUserServiceGetUserByNameCall) Do(f func(context.Context, user.Name) (user.User, error)) *MockUserServiceGetUserByNameCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockUserServiceGetUserByNameCall) DoAndReturn(f func(context.Context, user.Name) (user.User, error)) *MockUserServiceGetUserByNameCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
