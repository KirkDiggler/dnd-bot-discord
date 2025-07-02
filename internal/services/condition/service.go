package condition

import (
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/conditions"
	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
)

// Service manages conditions for all entities (characters, monsters, objects, etc.)
type Service interface {
	// AddCondition applies a condition to an entity
	AddCondition(entityID string, condType conditions.ConditionType, source string, duration conditions.DurationType, durationValue int) (*conditions.Condition, error)

	// RemoveCondition removes a specific condition
	RemoveCondition(entityID, conditionID string) error

	// RemoveConditionByType removes all conditions of a type from an entity
	RemoveConditionByType(entityID string, condType conditions.ConditionType) error

	// GetConditions returns all conditions for an entity
	GetConditions(entityID string) []*conditions.Condition

	// HasCondition checks if an entity has a specific condition type
	HasCondition(entityID string, condType conditions.ConditionType) bool

	// GetActiveEffects returns combined effects for an entity
	GetActiveEffects(entityID string) *conditions.Effect

	// ProcessTurnStart handles turn-based duration for an entity
	ProcessTurnStart(entityID string)

	// ProcessTurnEnd handles end-of-turn saves for an entity
	ProcessTurnEnd(entityID string, saveResults map[string]bool)

	// ProcessRoundEnd handles round-based duration for an entity
	ProcessRoundEnd(entityID string)

	// ProcessDamage handles conditions that end on damage
	ProcessDamage(entityID string, damageAmount int)
}

type service struct {
	mu       sync.RWMutex
	managers map[string]*conditions.Manager // entityID -> Manager
	eventBus *events.EventBus
}

// NewService creates a new condition service
func NewService(eventBus *events.EventBus) Service {
	svc := &service{
		managers: make(map[string]*conditions.Manager),
		eventBus: eventBus,
	}

	// Note: We don't register event handlers here to avoid circular dependencies.
	// Conditions should be applied directly through the service methods, not through events.
	// The service will emit events when conditions are applied/removed for other systems to react to.

	return svc
}

// getOrCreateManager gets or creates a condition manager for an entity
func (s *service) getOrCreateManager(entityID string) *conditions.Manager {
	s.mu.Lock()
	defer s.mu.Unlock()

	if manager, exists := s.managers[entityID]; exists {
		return manager
	}

	manager := conditions.NewManager(entityID, s.eventBus)
	s.managers[entityID] = manager
	return manager
}

func (s *service) AddCondition(entityID string, condType conditions.ConditionType, source string, duration conditions.DurationType, durationValue int) (*conditions.Condition, error) {
	manager := s.getOrCreateManager(entityID)
	return manager.AddCondition(condType, source, duration, durationValue)
}

func (s *service) RemoveCondition(entityID, conditionID string) error {
	manager := s.getOrCreateManager(entityID)
	return manager.RemoveCondition(conditionID)
}

func (s *service) RemoveConditionByType(entityID string, condType conditions.ConditionType) error {
	manager := s.getOrCreateManager(entityID)
	return manager.RemoveConditionByType(condType)
}

func (s *service) GetConditions(entityID string) []*conditions.Condition {
	manager := s.getOrCreateManager(entityID)
	return manager.GetConditions()
}

func (s *service) HasCondition(entityID string, condType conditions.ConditionType) bool {
	manager := s.getOrCreateManager(entityID)
	return manager.HasCondition(condType)
}

func (s *service) GetActiveEffects(entityID string) *conditions.Effect {
	manager := s.getOrCreateManager(entityID)
	return manager.GetActiveEffects()
}

func (s *service) ProcessTurnStart(entityID string) {
	manager := s.getOrCreateManager(entityID)
	manager.ProcessTurnStart()
}

func (s *service) ProcessTurnEnd(entityID string, saveResults map[string]bool) {
	manager := s.getOrCreateManager(entityID)
	manager.ProcessTurnEnd(saveResults)
}

func (s *service) ProcessRoundEnd(entityID string) {
	manager := s.getOrCreateManager(entityID)
	manager.ProcessRoundEnd()
}

func (s *service) ProcessDamage(entityID string, damageAmount int) {
	manager := s.getOrCreateManager(entityID)
	manager.ProcessDamage(damageAmount)
}
