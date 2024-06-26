// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/apiserver/common (interfaces: UnitStateBackend,UnitStateUnit)
//
// Generated by this command:
//
//	mockgen -typed -package mocks -destination mocks/unitstate.go github.com/juju/juju/apiserver/common UnitStateBackend,UnitStateUnit
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	common "github.com/juju/juju/apiserver/common"
	state "github.com/juju/juju/state"
	gomock "go.uber.org/mock/gomock"
)

// MockUnitStateBackend is a mock of UnitStateBackend interface.
type MockUnitStateBackend struct {
	ctrl     *gomock.Controller
	recorder *MockUnitStateBackendMockRecorder
}

// MockUnitStateBackendMockRecorder is the mock recorder for MockUnitStateBackend.
type MockUnitStateBackendMockRecorder struct {
	mock *MockUnitStateBackend
}

// NewMockUnitStateBackend creates a new mock instance.
func NewMockUnitStateBackend(ctrl *gomock.Controller) *MockUnitStateBackend {
	mock := &MockUnitStateBackend{ctrl: ctrl}
	mock.recorder = &MockUnitStateBackendMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnitStateBackend) EXPECT() *MockUnitStateBackendMockRecorder {
	return m.recorder
}

// ApplyOperation mocks base method.
func (m *MockUnitStateBackend) ApplyOperation(arg0 state.ModelOperation) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ApplyOperation", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// ApplyOperation indicates an expected call of ApplyOperation.
func (mr *MockUnitStateBackendMockRecorder) ApplyOperation(arg0 any) *MockUnitStateBackendApplyOperationCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ApplyOperation", reflect.TypeOf((*MockUnitStateBackend)(nil).ApplyOperation), arg0)
	return &MockUnitStateBackendApplyOperationCall{Call: call}
}

// MockUnitStateBackendApplyOperationCall wrap *gomock.Call
type MockUnitStateBackendApplyOperationCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockUnitStateBackendApplyOperationCall) Return(arg0 error) *MockUnitStateBackendApplyOperationCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockUnitStateBackendApplyOperationCall) Do(f func(state.ModelOperation) error) *MockUnitStateBackendApplyOperationCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockUnitStateBackendApplyOperationCall) DoAndReturn(f func(state.ModelOperation) error) *MockUnitStateBackendApplyOperationCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// Unit mocks base method.
func (m *MockUnitStateBackend) Unit(arg0 string) (common.UnitStateUnit, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Unit", arg0)
	ret0, _ := ret[0].(common.UnitStateUnit)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Unit indicates an expected call of Unit.
func (mr *MockUnitStateBackendMockRecorder) Unit(arg0 any) *MockUnitStateBackendUnitCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unit", reflect.TypeOf((*MockUnitStateBackend)(nil).Unit), arg0)
	return &MockUnitStateBackendUnitCall{Call: call}
}

// MockUnitStateBackendUnitCall wrap *gomock.Call
type MockUnitStateBackendUnitCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockUnitStateBackendUnitCall) Return(arg0 common.UnitStateUnit, arg1 error) *MockUnitStateBackendUnitCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockUnitStateBackendUnitCall) Do(f func(string) (common.UnitStateUnit, error)) *MockUnitStateBackendUnitCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockUnitStateBackendUnitCall) DoAndReturn(f func(string) (common.UnitStateUnit, error)) *MockUnitStateBackendUnitCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// MockUnitStateUnit is a mock of UnitStateUnit interface.
type MockUnitStateUnit struct {
	ctrl     *gomock.Controller
	recorder *MockUnitStateUnitMockRecorder
}

// MockUnitStateUnitMockRecorder is the mock recorder for MockUnitStateUnit.
type MockUnitStateUnitMockRecorder struct {
	mock *MockUnitStateUnit
}

// NewMockUnitStateUnit creates a new mock instance.
func NewMockUnitStateUnit(ctrl *gomock.Controller) *MockUnitStateUnit {
	mock := &MockUnitStateUnit{ctrl: ctrl}
	mock.recorder = &MockUnitStateUnitMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUnitStateUnit) EXPECT() *MockUnitStateUnitMockRecorder {
	return m.recorder
}

// SetStateOperation mocks base method.
func (m *MockUnitStateUnit) SetStateOperation(arg0 *state.UnitState, arg1 state.UnitStateSizeLimits) state.ModelOperation {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetStateOperation", arg0, arg1)
	ret0, _ := ret[0].(state.ModelOperation)
	return ret0
}

// SetStateOperation indicates an expected call of SetStateOperation.
func (mr *MockUnitStateUnitMockRecorder) SetStateOperation(arg0, arg1 any) *MockUnitStateUnitSetStateOperationCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetStateOperation", reflect.TypeOf((*MockUnitStateUnit)(nil).SetStateOperation), arg0, arg1)
	return &MockUnitStateUnitSetStateOperationCall{Call: call}
}

// MockUnitStateUnitSetStateOperationCall wrap *gomock.Call
type MockUnitStateUnitSetStateOperationCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockUnitStateUnitSetStateOperationCall) Return(arg0 state.ModelOperation) *MockUnitStateUnitSetStateOperationCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockUnitStateUnitSetStateOperationCall) Do(f func(*state.UnitState, state.UnitStateSizeLimits) state.ModelOperation) *MockUnitStateUnitSetStateOperationCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockUnitStateUnitSetStateOperationCall) DoAndReturn(f func(*state.UnitState, state.UnitStateSizeLimits) state.ModelOperation) *MockUnitStateUnitSetStateOperationCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// State mocks base method.
func (m *MockUnitStateUnit) State() (*state.UnitState, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "State")
	ret0, _ := ret[0].(*state.UnitState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// State indicates an expected call of State.
func (mr *MockUnitStateUnitMockRecorder) State() *MockUnitStateUnitStateCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "State", reflect.TypeOf((*MockUnitStateUnit)(nil).State))
	return &MockUnitStateUnitStateCall{Call: call}
}

// MockUnitStateUnitStateCall wrap *gomock.Call
type MockUnitStateUnitStateCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockUnitStateUnitStateCall) Return(arg0 *state.UnitState, arg1 error) *MockUnitStateUnitStateCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockUnitStateUnitStateCall) Do(f func() (*state.UnitState, error)) *MockUnitStateUnitStateCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockUnitStateUnitStateCall) DoAndReturn(f func() (*state.UnitState, error)) *MockUnitStateUnitStateCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}
