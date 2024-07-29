// Code generated by MockGen. DO NOT EDIT.
// Source: manager.go
//
// Generated by this command:
//
//	mockgen -source manager.go -destination manager_mock.go -package input
//

// Package input is a generated GoMock package.
package input

import (
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockInputManager is a mock of InputManager interface.
type MockInputManager struct {
	ctrl     *gomock.Controller
	recorder *MockInputManagerMockRecorder
}

// MockInputManagerMockRecorder is the mock recorder for MockInputManager.
type MockInputManagerMockRecorder struct {
	mock *MockInputManager
}

// NewMockInputManager creates a new mock instance.
func NewMockInputManager(ctrl *gomock.Controller) *MockInputManager {
	mock := &MockInputManager{ctrl: ctrl}
	mock.recorder = &MockInputManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInputManager) EXPECT() *MockInputManagerMockRecorder {
	return m.recorder
}

// GetInputResources mocks base method.
func (m *MockInputManager) GetInputResources(arg0 string) (*InputResource, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInputResources", arg0)
	ret0, _ := ret[0].(*InputResource)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetInputResources indicates an expected call of GetInputResources.
func (mr *MockInputManagerMockRecorder) GetInputResources(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInputResources", reflect.TypeOf((*MockInputManager)(nil).GetInputResources), arg0)
}

// InitInputs mocks base method.
func (m *MockInputManager) InitInputs() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InitInputs")
	ret0, _ := ret[0].(error)
	return ret0
}

// InitInputs indicates an expected call of InitInputs.
func (mr *MockInputManagerMockRecorder) InitInputs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InitInputs", reflect.TypeOf((*MockInputManager)(nil).InitInputs))
}

// StartInputs mocks base method.
func (m *MockInputManager) StartInputs() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StartInputs")
}

// StartInputs indicates an expected call of StartInputs.
func (mr *MockInputManagerMockRecorder) StartInputs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StartInputs", reflect.TypeOf((*MockInputManager)(nil).StartInputs))
}

// StopInputs mocks base method.
func (m *MockInputManager) StopInputs() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "StopInputs")
}

// StopInputs indicates an expected call of StopInputs.
func (mr *MockInputManagerMockRecorder) StopInputs() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "StopInputs", reflect.TypeOf((*MockInputManager)(nil).StopInputs))
}
