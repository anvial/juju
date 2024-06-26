// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/core/database (interfaces: DBGetter)
//
// Generated by this command:
//
//	mockgen -typed -package upgradedatabase -destination database_mock_test.go github.com/juju/juju/core/database DBGetter
//

// Package upgradedatabase is a generated GoMock package.
package upgradedatabase

import (
	reflect "reflect"

	database "github.com/juju/juju/core/database"
	gomock "go.uber.org/mock/gomock"
)

// MockDBGetter is a mock of DBGetter interface.
type MockDBGetter struct {
	ctrl     *gomock.Controller
	recorder *MockDBGetterMockRecorder
}

// MockDBGetterMockRecorder is the mock recorder for MockDBGetter.
type MockDBGetterMockRecorder struct {
	mock *MockDBGetter
}

// NewMockDBGetter creates a new mock instance.
func NewMockDBGetter(ctrl *gomock.Controller) *MockDBGetter {
	mock := &MockDBGetter{ctrl: ctrl}
	mock.recorder = &MockDBGetterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockDBGetter) EXPECT() *MockDBGetterMockRecorder {
	return m.recorder
}

// GetDB mocks base method.
func (m *MockDBGetter) GetDB(arg0 string) (database.TxnRunner, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDB", arg0)
	ret0, _ := ret[0].(database.TxnRunner)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetDB indicates an expected call of GetDB.
func (mr *MockDBGetterMockRecorder) GetDB(arg0 any) *MockDBGetterGetDBCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDB", reflect.TypeOf((*MockDBGetter)(nil).GetDB), arg0)
	return &MockDBGetterGetDBCall{Call: call}
}

// MockDBGetterGetDBCall wrap *gomock.Call
type MockDBGetterGetDBCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockDBGetterGetDBCall) Return(arg0 database.TxnRunner, arg1 error) *MockDBGetterGetDBCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockDBGetterGetDBCall) Do(f func(string) (database.TxnRunner, error)) *MockDBGetterGetDBCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockDBGetterGetDBCall) DoAndReturn(f func(string) (database.TxnRunner, error)) *MockDBGetterGetDBCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
