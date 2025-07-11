// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e (interfaces: Client)
//
// Generated by this command:
//
//	mockgen -destination=internal/mocks/dnd5e/mock_client.go -package=mockdnd5e github.com/KirkDiggler/dnd-bot-discord/internal/clients/dnd5e Client
//

// Package mockdnd5e is a generated GoMock package.
package mockdnd5e

import (
	reflect "reflect"

	equipment "github.com/KirkDiggler/dnd-bot-discord/internal/domain/equipment"
	combat "github.com/KirkDiggler/dnd-bot-discord/internal/domain/game/combat"
	rulebook "github.com/KirkDiggler/dnd-bot-discord/internal/domain/rulebook/dnd5e"
	gomock "go.uber.org/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
	isgomock struct{}
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// GetClass mocks base method.
func (m *MockClient) GetClass(key string) (*rulebook.Class, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClass", key)
	ret0, _ := ret[0].(*rulebook.Class)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClass indicates an expected call of GetClass.
func (mr *MockClientMockRecorder) GetClass(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClass", reflect.TypeOf((*MockClient)(nil).GetClass), key)
}

// GetClassFeatures mocks base method.
func (m *MockClient) GetClassFeatures(classKey string, level int) ([]*rulebook.CharacterFeature, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetClassFeatures", classKey, level)
	ret0, _ := ret[0].([]*rulebook.CharacterFeature)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetClassFeatures indicates an expected call of GetClassFeatures.
func (mr *MockClientMockRecorder) GetClassFeatures(classKey, level any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetClassFeatures", reflect.TypeOf((*MockClient)(nil).GetClassFeatures), classKey, level)
}

// GetEquipment mocks base method.
func (m *MockClient) GetEquipment(key string) (equipment.Equipment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEquipment", key)
	ret0, _ := ret[0].(equipment.Equipment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEquipment indicates an expected call of GetEquipment.
func (mr *MockClientMockRecorder) GetEquipment(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEquipment", reflect.TypeOf((*MockClient)(nil).GetEquipment), key)
}

// GetEquipmentByCategory mocks base method.
func (m *MockClient) GetEquipmentByCategory(category string) ([]equipment.Equipment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetEquipmentByCategory", category)
	ret0, _ := ret[0].([]equipment.Equipment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetEquipmentByCategory indicates an expected call of GetEquipmentByCategory.
func (mr *MockClientMockRecorder) GetEquipmentByCategory(category any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetEquipmentByCategory", reflect.TypeOf((*MockClient)(nil).GetEquipmentByCategory), category)
}

// GetMonster mocks base method.
func (m *MockClient) GetMonster(key string) (*combat.MonsterTemplate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetMonster", key)
	ret0, _ := ret[0].(*combat.MonsterTemplate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetMonster indicates an expected call of GetMonster.
func (mr *MockClientMockRecorder) GetMonster(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetMonster", reflect.TypeOf((*MockClient)(nil).GetMonster), key)
}

// GetProficiency mocks base method.
func (m *MockClient) GetProficiency(key string) (*rulebook.Proficiency, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProficiency", key)
	ret0, _ := ret[0].(*rulebook.Proficiency)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProficiency indicates an expected call of GetProficiency.
func (mr *MockClientMockRecorder) GetProficiency(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProficiency", reflect.TypeOf((*MockClient)(nil).GetProficiency), key)
}

// GetRace mocks base method.
func (m *MockClient) GetRace(key string) (*rulebook.Race, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRace", key)
	ret0, _ := ret[0].(*rulebook.Race)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetRace indicates an expected call of GetRace.
func (mr *MockClientMockRecorder) GetRace(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRace", reflect.TypeOf((*MockClient)(nil).GetRace), key)
}

// GetSpell mocks base method.
func (m *MockClient) GetSpell(key string) (*rulebook.Spell, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetSpell", key)
	ret0, _ := ret[0].(*rulebook.Spell)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetSpell indicates an expected call of GetSpell.
func (mr *MockClientMockRecorder) GetSpell(key any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetSpell", reflect.TypeOf((*MockClient)(nil).GetSpell), key)
}

// ListClasses mocks base method.
func (m *MockClient) ListClasses() ([]*rulebook.Class, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListClasses")
	ret0, _ := ret[0].([]*rulebook.Class)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListClasses indicates an expected call of ListClasses.
func (mr *MockClientMockRecorder) ListClasses() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListClasses", reflect.TypeOf((*MockClient)(nil).ListClasses))
}

// ListEquipment mocks base method.
func (m *MockClient) ListEquipment() ([]equipment.Equipment, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListEquipment")
	ret0, _ := ret[0].([]equipment.Equipment)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListEquipment indicates an expected call of ListEquipment.
func (mr *MockClientMockRecorder) ListEquipment() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListEquipment", reflect.TypeOf((*MockClient)(nil).ListEquipment))
}

// ListMonstersByCR mocks base method.
func (m *MockClient) ListMonstersByCR(minCR, maxCR float32) ([]*combat.MonsterTemplate, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListMonstersByCR", minCR, maxCR)
	ret0, _ := ret[0].([]*combat.MonsterTemplate)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListMonstersByCR indicates an expected call of ListMonstersByCR.
func (mr *MockClientMockRecorder) ListMonstersByCR(minCR, maxCR any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListMonstersByCR", reflect.TypeOf((*MockClient)(nil).ListMonstersByCR), minCR, maxCR)
}

// ListRaces mocks base method.
func (m *MockClient) ListRaces() ([]*rulebook.Race, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListRaces")
	ret0, _ := ret[0].([]*rulebook.Race)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListRaces indicates an expected call of ListRaces.
func (mr *MockClientMockRecorder) ListRaces() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListRaces", reflect.TypeOf((*MockClient)(nil).ListRaces))
}

// ListSpellsByClass mocks base method.
func (m *MockClient) ListSpellsByClass(classKey string) ([]*rulebook.SpellReference, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSpellsByClass", classKey)
	ret0, _ := ret[0].([]*rulebook.SpellReference)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSpellsByClass indicates an expected call of ListSpellsByClass.
func (mr *MockClientMockRecorder) ListSpellsByClass(classKey any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSpellsByClass", reflect.TypeOf((*MockClient)(nil).ListSpellsByClass), classKey)
}

// ListSpellsByClassAndLevel mocks base method.
func (m *MockClient) ListSpellsByClassAndLevel(classKey string, level int) ([]*rulebook.SpellReference, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListSpellsByClassAndLevel", classKey, level)
	ret0, _ := ret[0].([]*rulebook.SpellReference)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListSpellsByClassAndLevel indicates an expected call of ListSpellsByClassAndLevel.
func (mr *MockClientMockRecorder) ListSpellsByClassAndLevel(classKey, level any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListSpellsByClassAndLevel", reflect.TypeOf((*MockClient)(nil).ListSpellsByClassAndLevel), classKey, level)
}
