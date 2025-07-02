package conditions

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/KirkDiggler/dnd-bot-discord/internal/domain/events"
	"github.com/KirkDiggler/dnd-bot-discord/internal/uuid"
)

// Manager handles condition tracking for an entity (character, monster, object, etc.)
type Manager struct {
	mu            sync.RWMutex
	conditions    map[string]*Condition // Key is condition ID
	entityID      string
	eventBus      *events.EventBus
	uuidGenerator uuid.Generator
}

// NewManager creates a new condition manager
func NewManager(entityID string, eventBus *events.EventBus) *Manager {
	return &Manager{
		conditions:    make(map[string]*Condition),
		entityID:      entityID,
		eventBus:      eventBus,
		uuidGenerator: uuid.NewGoogleUUIDGenerator(),
	}
}

// AddCondition applies a new condition to the character
func (m *Manager) AddCondition(condType ConditionType, source string, duration DurationType, durationValue int) (*Condition, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if we already have this condition type
	// Some conditions don't stack (like poisoned), others might (exhaustion levels)
	if existingCond := m.getConditionByType(condType); existingCond != nil {
		switch condType {
		case Exhaustion:
			// Exhaustion stacks up to level 6
			if existingCond.Level < 6 {
				existingCond.Level++
				m.emitConditionModified(existingCond)
				return existingCond, nil
			}
			return existingCond, fmt.Errorf("already at maximum exhaustion level")
		default:
			// Most conditions don't stack - refresh duration if new is longer
			if duration == DurationRounds && durationValue > existingCond.Remaining {
				existingCond.Remaining = durationValue
				existingCond.Duration = durationValue
				m.emitConditionModified(existingCond)
			}
			return existingCond, nil
		}
	}

	// Create new condition
	condition := &Condition{
		ID:           m.uuidGenerator.New(),
		Type:         condType,
		Name:         string(condType), // Could be enhanced with proper names
		Source:       source,
		DurationType: duration,
		Duration:     durationValue,
		Remaining:    durationValue,
		AppliedAt:    time.Now(),
		Metadata:     make(map[string]interface{}),
	}

	// Set level for exhaustion
	if condType == Exhaustion {
		condition.Level = 1
	}

	// Add standard description
	condition.Description = m.getConditionDescription(condType)

	// Store condition
	m.conditions[condition.ID] = condition

	// Emit event
	m.emitConditionApplied(condition)

	log.Printf("[CONDITIONS] Applied %s to entity %s (duration: %s for %d)",
		condType, m.entityID, duration, durationValue)

	return condition, nil
}

// RemoveCondition removes a condition by ID
func (m *Manager) RemoveCondition(conditionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	condition, exists := m.conditions[conditionID]
	if !exists {
		return fmt.Errorf("condition %s not found", conditionID)
	}

	delete(m.conditions, conditionID)
	m.emitConditionRemoved(condition)

	log.Printf("[CONDITIONS] Removed %s from entity %s",
		condition.Type, m.entityID)

	return nil
}

// RemoveConditionByType removes all conditions of a specific type
func (m *Manager) RemoveConditionByType(condType ConditionType) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var removed []*Condition
	for id, cond := range m.conditions {
		if cond.Type == condType {
			delete(m.conditions, id)
			removed = append(removed, cond)
		}
	}

	for _, cond := range removed {
		m.emitConditionRemoved(cond)
	}

	return nil
}

// GetConditions returns all active conditions
func (m *Manager) GetConditions() []*Condition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conditions := make([]*Condition, 0, len(m.conditions))
	for _, cond := range m.conditions {
		conditions = append(conditions, cond)
	}
	return conditions
}

// HasCondition checks if a character has a specific condition type
func (m *Manager) HasCondition(condType ConditionType) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getConditionByType(condType) != nil
}

// GetConditionByType returns the active condition of a specific type
func (m *Manager) GetConditionByType(condType ConditionType) *Condition {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getConditionByType(condType)
}

// ProcessTurnStart handles condition duration at start of turn
func (m *Manager) ProcessTurnStart() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var toRemove []string

	for id, cond := range m.conditions {
		// Handle turn-based durations
		if cond.DurationType == DurationTurns {
			if cond.Remaining > 0 {
				cond.Remaining--
			} else if cond.Remaining == 0 {
				// Expired, remove it
				toRemove = append(toRemove, id)
			}
		}
	}

	// Remove expired conditions
	for _, id := range toRemove {
		if cond, exists := m.conditions[id]; exists {
			delete(m.conditions, id)
			m.emitConditionRemoved(cond)
			log.Printf("[CONDITIONS] %s expired on entity %s", cond.Type, m.entityID)
		}
	}
}

// ProcessTurnEnd handles save attempts at end of turn
func (m *Manager) ProcessTurnEnd(saveResults map[string]bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var toRemove []string

	for id, cond := range m.conditions {
		// Check for end-of-turn saves
		if cond.SaveEnd && cond.SaveDC > 0 {
			// Check if a save was made
			if saved, hasSave := saveResults[string(cond.Type)]; hasSave && saved {
				toRemove = append(toRemove, id)
				log.Printf("[CONDITIONS] Entity %s saved against %s", m.entityID, cond.Type)
			}
		}

		// Handle "end of next turn" durations
		if cond.DurationType == DurationEndOfNextTurn {
			toRemove = append(toRemove, id)
		}
	}

	// Remove conditions that were saved against or expired
	for _, id := range toRemove {
		if cond, exists := m.conditions[id]; exists {
			delete(m.conditions, id)
			m.emitConditionRemoved(cond)
		}
	}
}

// ProcessRoundEnd handles round-based duration
func (m *Manager) ProcessRoundEnd() {
	m.mu.Lock()
	defer m.mu.Unlock()

	var toRemove []string

	for id, cond := range m.conditions {
		// Handle round-based durations
		if cond.DurationType == DurationRounds {
			if cond.Remaining > 0 {
				cond.Remaining--
			} else if cond.Remaining == 0 {
				// Expired, remove it
				toRemove = append(toRemove, id)
			}
		}
	}

	// Remove expired conditions
	for _, id := range toRemove {
		if cond, exists := m.conditions[id]; exists {
			delete(m.conditions, id)
			m.emitConditionRemoved(cond)
			log.Printf("[CONDITIONS] %s expired on entity %s", cond.Type, m.entityID)
		}
	}
}

// ProcessDamage handles conditions that end when damaged
func (m *Manager) ProcessDamage(damageAmount int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var toRemove []string

	for id, cond := range m.conditions {
		// Remove conditions that end on damage
		if cond.DurationType == DurationUntilDamaged && damageAmount > 0 {
			toRemove = append(toRemove, id)
		}

		// Concentration checks would also go here
		if cond.Type == Concentration && damageAmount > 0 {
			// TODO: Trigger concentration save
			// For now, we'll emit an event for the save system to handle
			if m.eventBus != nil {
				concEvent := events.NewGameEvent(events.OnConcentrationCheck).
					WithContext("entity_id", m.entityID).
					WithContext("damage", damageAmount).
					WithContext("condition_id", id)
				if err := m.eventBus.Emit(concEvent); err != nil {
					log.Printf("Failed to emit OnConcentrationEnded event: %v", err)
				}
			}
		}
	}

	// Remove conditions that end on damage
	for _, id := range toRemove {
		if cond, exists := m.conditions[id]; exists {
			delete(m.conditions, id)
			m.emitConditionRemoved(cond)
			log.Printf("[CONDITIONS] %s ended on entity %s due to damage", cond.Type, m.entityID)
		}
	}
}

// GetActiveEffects compiles all effects from active conditions
func (m *Manager) GetActiveEffects() *Effect {
	m.mu.RLock()
	defer m.mu.RUnlock()

	combined := &Effect{
		SaveAdvantage:    make(map[string]bool),
		SaveDisadvantage: make(map[string]bool),
		SaveAutoFail:     make(map[string]bool),
		Vulnerability:    make(map[string]bool),
		Resistance:       make(map[string]bool),
		Immunity:         make(map[string]bool),
	}

	for _, cond := range m.conditions {
		effect := GetStandardEffects(cond.Type)
		m.mergeEffects(combined, effect)
	}

	return combined
}

// Helper methods

func (m *Manager) getConditionByType(condType ConditionType) *Condition {
	for _, cond := range m.conditions {
		if cond.Type == condType {
			return cond
		}
	}
	return nil
}

func (m *Manager) mergeEffects(target, source *Effect) {
	// Merge boolean flags (OR operation - if any is true, result is true)
	target.AttackAdvantage = target.AttackAdvantage || source.AttackAdvantage
	target.AttackDisadvantage = target.AttackDisadvantage || source.AttackDisadvantage
	target.DefenseAdvantage = target.DefenseAdvantage || source.DefenseAdvantage
	target.DefenseDisadvantage = target.DefenseDisadvantage || source.DefenseDisadvantage
	target.CantMove = target.CantMove || source.CantMove
	target.CantAct = target.CantAct || source.CantAct
	target.CantReact = target.CantReact || source.CantReact
	target.CantSpeak = target.CantSpeak || source.CantSpeak
	target.Incapacitated = target.Incapacitated || source.Incapacitated
	target.FallProne = target.FallProne || source.FallProne
	target.DropItems = target.DropItems || source.DropItems
	target.CantConcentrate = target.CantConcentrate || source.CantConcentrate

	// Merge speed multiplier (take the lowest)
	if source.SpeedMultiplier > 0 && (target.SpeedMultiplier == 0 || source.SpeedMultiplier < target.SpeedMultiplier) {
		target.SpeedMultiplier = source.SpeedMultiplier
	}

	// Merge maps
	for k, v := range source.SaveAdvantage {
		if v {
			target.SaveAdvantage[k] = v
		}
	}
	for k, v := range source.SaveDisadvantage {
		if v {
			target.SaveDisadvantage[k] = v
		}
	}
	for k, v := range source.SaveAutoFail {
		if v {
			target.SaveAutoFail[k] = v
		}
	}
	for k, v := range source.Vulnerability {
		if v {
			target.Vulnerability[k] = v
		}
	}
	for k, v := range source.Resistance {
		if v {
			target.Resistance[k] = v
		}
	}
	for k, v := range source.Immunity {
		if v {
			target.Immunity[k] = v
		}
	}
}

func (m *Manager) getConditionDescription(condType ConditionType) string {
	descriptions := map[ConditionType]string{
		Blinded:                  "Can't see. Auto-fails sight checks. Disadvantage on attacks. Attacks against have advantage.",
		Charmed:                  "Can't attack charmer. Charmer has advantage on social checks.",
		Deafened:                 "Can't hear. Auto-fails hearing checks.",
		Frightened:               "Disadvantage on ability checks and attacks while source is in sight. Can't willingly move closer.",
		Grappled:                 "Speed is 0. Condition ends if grappler is incapacitated or moved away.",
		Incapacitated:            "Can't take actions or reactions.",
		Invisible:                "Can't be seen. Has advantage on attacks. Attacks against have disadvantage.",
		Paralyzed:                "Incapacitated, can't move or speak. Auto-fails STR and DEX saves. Attacks have advantage. Hits within 5 ft are crits.",
		Petrified:                "Transformed to stone. Weight x10. Incapacitated, can't move or speak. Resistance to all damage. Immune to poison/disease.",
		Poisoned:                 "Disadvantage on attack rolls and ability checks.",
		Prone:                    "Only move by crawling. Disadvantage on attacks. Melee attacks within 5 ft have advantage, ranged have disadvantage.",
		Restrained:               "Speed is 0. Disadvantage on attacks and DEX saves. Attacks against have advantage.",
		Stunned:                  "Incapacitated, can't move, can speak only falteringly. Auto-fails STR and DEX saves. Attacks against have advantage.",
		Unconscious:              "Incapacitated, can't move or speak, unaware. Drops held items and falls prone. Auto-fails STR and DEX saves. Attacks have advantage. Hits within 5 ft are crits.",
		Exhaustion:               "Level-based penalties. 1: Disadvantage on checks. 2: Speed halved. 3: Disadvantage on attacks/saves. 4: HP max halved. 5: Speed is 0. 6: Death.",
		DisadvantageOnNextAttack: "The next attack roll has disadvantage.",
		Concentration:            "Concentrating on a spell. Taking damage requires a save or lose concentration.",
		Rage:                     "Advantage on STR checks/saves. Bonus damage. Resistance to physical damage.",
	}

	if desc, exists := descriptions[condType]; exists {
		return desc
	}
	return "Unknown condition effect."
}

// Event emission helpers

func (m *Manager) emitConditionApplied(condition *Condition) {
	if m.eventBus != nil {
		event := events.NewGameEvent(events.OnConditionApplied).
			WithContext("entity_id", m.entityID).
			WithContext("condition_id", condition.ID).
			WithContext("condition_type", string(condition.Type)).
			WithContext("source", condition.Source)
		if err := m.eventBus.Emit(event); err != nil {
			log.Printf("Failed to emit event: %v", err)
		}
	}
}

func (m *Manager) emitConditionRemoved(condition *Condition) {
	if m.eventBus != nil {
		event := events.NewGameEvent(events.OnConditionRemoved).
			WithContext("entity_id", m.entityID).
			WithContext("condition_id", condition.ID).
			WithContext("condition_type", string(condition.Type))
		if err := m.eventBus.Emit(event); err != nil {
			log.Printf("Failed to emit event: %v", err)
		}
	}
}

func (m *Manager) emitConditionModified(condition *Condition) {
	if m.eventBus != nil {
		event := events.NewGameEvent(events.OnConditionModified).
			WithContext("entity_id", m.entityID).
			WithContext("condition_id", condition.ID).
			WithContext("condition_type", string(condition.Type)).
			WithContext("level", condition.Level)
		if err := m.eventBus.Emit(event); err != nil {
			log.Printf("Failed to emit event: %v", err)
		}
	}
}
