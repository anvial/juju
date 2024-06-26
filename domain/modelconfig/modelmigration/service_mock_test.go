// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/domain/modelconfig/service (interfaces: ModelDefaultsProvider)
//
// Generated by this command:
//
//	mockgen -typed -package modelmigration -destination service_mock_test.go github.com/juju/juju/domain/modelconfig/service ModelDefaultsProvider
//

// Package modelmigration is a generated GoMock package.
package modelmigration

import (
	context "context"
	reflect "reflect"

	modeldefaults "github.com/juju/juju/domain/modeldefaults"
	gomock "go.uber.org/mock/gomock"
)

// MockModelDefaultsProvider is a mock of ModelDefaultsProvider interface.
type MockModelDefaultsProvider struct {
	ctrl     *gomock.Controller
	recorder *MockModelDefaultsProviderMockRecorder
}

// MockModelDefaultsProviderMockRecorder is the mock recorder for MockModelDefaultsProvider.
type MockModelDefaultsProviderMockRecorder struct {
	mock *MockModelDefaultsProvider
}

// NewMockModelDefaultsProvider creates a new mock instance.
func NewMockModelDefaultsProvider(ctrl *gomock.Controller) *MockModelDefaultsProvider {
	mock := &MockModelDefaultsProvider{ctrl: ctrl}
	mock.recorder = &MockModelDefaultsProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockModelDefaultsProvider) EXPECT() *MockModelDefaultsProviderMockRecorder {
	return m.recorder
}

// ModelDefaults mocks base method.
func (m *MockModelDefaultsProvider) ModelDefaults(arg0 context.Context) (modeldefaults.Defaults, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ModelDefaults", arg0)
	ret0, _ := ret[0].(modeldefaults.Defaults)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ModelDefaults indicates an expected call of ModelDefaults.
func (mr *MockModelDefaultsProviderMockRecorder) ModelDefaults(arg0 any) *MockModelDefaultsProviderModelDefaultsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ModelDefaults", reflect.TypeOf((*MockModelDefaultsProvider)(nil).ModelDefaults), arg0)
	return &MockModelDefaultsProviderModelDefaultsCall{Call: call}
}

// MockModelDefaultsProviderModelDefaultsCall wrap *gomock.Call
type MockModelDefaultsProviderModelDefaultsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockModelDefaultsProviderModelDefaultsCall) Return(arg0 modeldefaults.Defaults, arg1 error) *MockModelDefaultsProviderModelDefaultsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockModelDefaultsProviderModelDefaultsCall) Do(f func(context.Context) (modeldefaults.Defaults, error)) *MockModelDefaultsProviderModelDefaultsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockModelDefaultsProviderModelDefaultsCall) DoAndReturn(f func(context.Context) (modeldefaults.Defaults, error)) *MockModelDefaultsProviderModelDefaultsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
