// Code generated by MockGen. DO NOT EDIT.
// Source: service.go
//
// Generated by this command:
//
//	mockgen -destination=mock/mock_service.go -package=mockloot -source=service.go
//

// Package mockloot is a generated GoMock package.
package mockloot

import (
	context "context"
	reflect "reflect"

	loot "github.com/KirkDiggler/dnd-bot-discord/internal/services/loot"
	gomock "go.uber.org/mock/gomock"
)

// MockService is a mock of Service interface.
type MockService struct {
	ctrl     *gomock.Controller
	recorder *MockServiceMockRecorder
	isgomock struct{}
}

// MockServiceMockRecorder is the mock recorder for MockService.
type MockServiceMockRecorder struct {
	mock *MockService
}

// NewMockService creates a new mock instance.
func NewMockService(ctrl *gomock.Controller) *MockService {
	mock := &MockService{ctrl: ctrl}
	mock.recorder = &MockServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockService) EXPECT() *MockServiceMockRecorder {
	return m.recorder
}

// GenerateLootTable mocks base method.
func (m *MockService) GenerateLootTable(ctx context.Context, challengeRating float64) (*loot.LootTable, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateLootTable", ctx, challengeRating)
	ret0, _ := ret[0].(*loot.LootTable)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateLootTable indicates an expected call of GenerateLootTable.
func (mr *MockServiceMockRecorder) GenerateLootTable(ctx, challengeRating any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateLootTable", reflect.TypeOf((*MockService)(nil).GenerateLootTable), ctx, challengeRating)
}

// GenerateTreasure mocks base method.
func (m *MockService) GenerateTreasure(ctx context.Context, difficulty string, roomNumber int) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateTreasure", ctx, difficulty, roomNumber)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateTreasure indicates an expected call of GenerateTreasure.
func (mr *MockServiceMockRecorder) GenerateTreasure(ctx, difficulty, roomNumber any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateTreasure", reflect.TypeOf((*MockService)(nil).GenerateTreasure), ctx, difficulty, roomNumber)
}
