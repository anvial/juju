// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/juju/juju/apiserver/facades/client/annotations (interfaces: AnnotationService)
//
// Generated by this command:
//
//	mockgen -typed -package annotations -destination annotations_mock_test.go github.com/juju/juju/apiserver/facades/client/annotations AnnotationService
//

// Package annotations is a generated GoMock package.
package annotations

import (
	context "context"
	reflect "reflect"

	annotations "github.com/juju/juju/core/annotations"
	gomock "go.uber.org/mock/gomock"
)

// MockAnnotationService is a mock of AnnotationService interface.
type MockAnnotationService struct {
	ctrl     *gomock.Controller
	recorder *MockAnnotationServiceMockRecorder
}

// MockAnnotationServiceMockRecorder is the mock recorder for MockAnnotationService.
type MockAnnotationServiceMockRecorder struct {
	mock *MockAnnotationService
}

// NewMockAnnotationService creates a new mock instance.
func NewMockAnnotationService(ctrl *gomock.Controller) *MockAnnotationService {
	mock := &MockAnnotationService{ctrl: ctrl}
	mock.recorder = &MockAnnotationServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAnnotationService) EXPECT() *MockAnnotationServiceMockRecorder {
	return m.recorder
}

// GetAnnotations mocks base method.
func (m *MockAnnotationService) GetAnnotations(arg0 context.Context, arg1 annotations.ID) (map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAnnotations", arg0, arg1)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAnnotations indicates an expected call of GetAnnotations.
func (mr *MockAnnotationServiceMockRecorder) GetAnnotations(arg0, arg1 any) *MockAnnotationServiceGetAnnotationsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAnnotations", reflect.TypeOf((*MockAnnotationService)(nil).GetAnnotations), arg0, arg1)
	return &MockAnnotationServiceGetAnnotationsCall{Call: call}
}

// MockAnnotationServiceGetAnnotationsCall wrap *gomock.Call
type MockAnnotationServiceGetAnnotationsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockAnnotationServiceGetAnnotationsCall) Return(arg0 map[string]string, arg1 error) *MockAnnotationServiceGetAnnotationsCall {
	c.Call = c.Call.Return(arg0, arg1)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockAnnotationServiceGetAnnotationsCall) Do(f func(context.Context, annotations.ID) (map[string]string, error)) *MockAnnotationServiceGetAnnotationsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockAnnotationServiceGetAnnotationsCall) DoAndReturn(f func(context.Context, annotations.ID) (map[string]string, error)) *MockAnnotationServiceGetAnnotationsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}

// SetAnnotations mocks base method.
func (m *MockAnnotationService) SetAnnotations(arg0 context.Context, arg1 annotations.ID, arg2 map[string]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetAnnotations", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetAnnotations indicates an expected call of SetAnnotations.
func (mr *MockAnnotationServiceMockRecorder) SetAnnotations(arg0, arg1, arg2 any) *MockAnnotationServiceSetAnnotationsCall {
	mr.mock.ctrl.T.Helper()
	call := mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetAnnotations", reflect.TypeOf((*MockAnnotationService)(nil).SetAnnotations), arg0, arg1, arg2)
	return &MockAnnotationServiceSetAnnotationsCall{Call: call}
}

// MockAnnotationServiceSetAnnotationsCall wrap *gomock.Call
type MockAnnotationServiceSetAnnotationsCall struct {
	*gomock.Call
}

// Return rewrite *gomock.Call.Return
func (c *MockAnnotationServiceSetAnnotationsCall) Return(arg0 error) *MockAnnotationServiceSetAnnotationsCall {
	c.Call = c.Call.Return(arg0)
	return c
}

// Do rewrite *gomock.Call.Do
func (c *MockAnnotationServiceSetAnnotationsCall) Do(f func(context.Context, annotations.ID, map[string]string) error) *MockAnnotationServiceSetAnnotationsCall {
	c.Call = c.Call.Do(f)
	return c
}

// DoAndReturn rewrite *gomock.Call.DoAndReturn
func (c *MockAnnotationServiceSetAnnotationsCall) DoAndReturn(f func(context.Context, annotations.ID, map[string]string) error) *MockAnnotationServiceSetAnnotationsCall {
	c.Call = c.Call.DoAndReturn(f)
	return c
}