// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/KirkDiggler/dnd-bot-discord/internal/repositories/sessions (interfaces: TimeProvider)
//
// Generated by this command:
//
//	mockgen -destination=mocks/mock_time_provider.go -package=mocks github.com/KirkDiggler/dnd-bot-discord/internal/repositories/sessions TimeProvider
//

// Package mocks is a generated GoMock package.
package mocks

import (
	reflect "reflect"
	time "time"

	gomock "go.uber.org/mock/gomock"
)

// MockTimeProvider is a mock of TimeProvider interface.
type MockTimeProvider struct {
	ctrl     *gomock.Controller
	recorder *MockTimeProviderMockRecorder
	isgomock struct{}
}

// MockTimeProviderMockRecorder is the mock recorder for MockTimeProvider.
type MockTimeProviderMockRecorder struct {
	mock *MockTimeProvider
}

// NewMockTimeProvider creates a new mock instance.
func NewMockTimeProvider(ctrl *gomock.Controller) *MockTimeProvider {
	mock := &MockTimeProvider{ctrl: ctrl}
	mock.recorder = &MockTimeProviderMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockTimeProvider) EXPECT() *MockTimeProviderMockRecorder {
	return m.recorder
}

// Now mocks base method.
func (m *MockTimeProvider) Now() time.Time {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Now")
	ret0, _ := ret[0].(time.Time)
	return ret0
}

// Now indicates an expected call of Now.
func (mr *MockTimeProviderMockRecorder) Now() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Now", reflect.TypeOf((*MockTimeProvider)(nil).Now))
}
