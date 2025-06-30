package modifiers

import (
	"fmt"
	"log"
	"sort"
	"sync"

	"github.com/KirkDiggler/dnd-bot-discord/internal/entities"
	"github.com/KirkDiggler/dnd-bot-discord/internal/events"
)

// Manager handles modifiers for a character
type Manager struct {
	characterID   string
	modifiers     []Modifier
	eventBus      *events.Bus
	subscriptions map[string][]events.EventType // modifier ID -> subscribed event types
	mu            sync.RWMutex
}

// NewManager creates a new modifier manager
func NewManager(characterID string, eventBus *events.Bus) *Manager {
	return &Manager{
		characterID:   characterID,
		modifiers:     make([]Modifier, 0),
		eventBus:      eventBus,
		subscriptions: make(map[string][]events.EventType),
	}
}

// AddModifier adds a modifier and registers it with the event system
func (m *Manager) AddModifier(mod Modifier) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if modifier already exists
	for i, existing := range m.modifiers {
		if existing.ID() == mod.ID() {
			// Replace existing
			m.modifiers[i] = mod
			log.Printf("ModifierManager: Replaced modifier %s", mod.ID())
			return
		}
	}

	// Add new modifier
	m.modifiers = append(m.modifiers, mod)

	// Sort by priority
	sort.Slice(m.modifiers, func(i, j int) bool {
		return m.modifiers[i].Priority() < m.modifiers[j].Priority()
	})

	log.Printf("ModifierManager: Added modifier %s from %s", mod.ID(), mod.Source().Name)
}

// RemoveModifier removes a modifier by ID
func (m *Manager) RemoveModifier(modifierID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Unsubscribe from events first
	if eventTypes, ok := m.subscriptions[modifierID]; ok {
		listenerID := fmt.Sprintf("modifier_%s", modifierID)
		for _, eventType := range eventTypes {
			m.eventBus.Unsubscribe(eventType, listenerID)
		}
		delete(m.subscriptions, modifierID)
	}

	for i, mod := range m.modifiers {
		if mod.ID() != modifierID {
			continue
		}
		// Remove by swapping with last and truncating
		m.modifiers[i] = m.modifiers[len(m.modifiers)-1]
		m.modifiers = m.modifiers[:len(m.modifiers)-1]

		// Re-sort
		sort.Slice(m.modifiers, func(i, j int) bool {
			return m.modifiers[i].Priority() < m.modifiers[j].Priority()
		})

		log.Printf("ModifierManager: Removed modifier %s", modifierID)
		return
	}
}

// GetActiveModifiers returns all currently active modifiers
func (m *Manager) GetActiveModifiers(character *entities.Character) []Modifier {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := make([]Modifier, 0)
	expired := make([]string, 0)

	for _, mod := range m.modifiers {
		if mod.Duration().IsExpired() {
			expired = append(expired, mod.ID())
			continue
		}

		if mod.IsActive(character) {
			active = append(active, mod)
		}
	}

	// Clean up expired modifiers (do this after releasing read lock)
	if len(expired) > 0 {
		m.mu.RUnlock()
		m.mu.Lock()
		// Remove expired modifiers directly to avoid nested locking
		for _, id := range expired {
			for i, mod := range m.modifiers {
				if mod.ID() == id {
					// Remove by swapping with last and truncating
					m.modifiers[i] = m.modifiers[len(m.modifiers)-1]
					m.modifiers = m.modifiers[:len(m.modifiers)-1]
					log.Printf("ModifierManager: Removed expired modifier %s", id)
					break
				}
			}
		}
		// Re-sort after removal
		sort.Slice(m.modifiers, func(i, j int) bool {
			return m.modifiers[i].Priority() < m.modifiers[j].Priority()
		})
		m.mu.Unlock()
		m.mu.RLock()
	}

	return active
}

// ModifierListener adapts modifiers to work with the event bus
type ModifierListener struct {
	modifier  Modifier
	character *entities.Character
	manager   *Manager
}

func (l *ModifierListener) ID() string {
	return fmt.Sprintf("modifier_%s", l.modifier.ID())
}

func (l *ModifierListener) Priority() int {
	return l.modifier.Priority()
}

func (l *ModifierListener) HandleEvent(event events.Event) error {
	// Check if modifier is still active
	if !l.modifier.IsActive(l.character) {
		return nil
	}

	// Let the modifier's duration track the event
	l.modifier.Duration().OnEventOccurred(event)

	// Apply the modifier's effects via visitor pattern
	event.Accept(l.modifier.AsVisitor())

	return nil
}

// CreateListener creates an event listener for a modifier
func (m *Manager) CreateListener(mod Modifier, character *entities.Character) events.EventListener {
	return &ModifierListener{
		modifier:  mod,
		character: character,
		manager:   m,
	}
}

// SubscribeAll subscribes all active modifiers to relevant events
func (m *Manager) SubscribeAll(character *entities.Character) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Unsubscribe all existing subscriptions first
	for modID, eventTypes := range m.subscriptions {
		listenerID := fmt.Sprintf("modifier_%s", modID)
		for _, eventType := range eventTypes {
			m.eventBus.Unsubscribe(eventType, listenerID)
		}
	}
	m.subscriptions = make(map[string][]events.EventType)

	// For now, subscribe all modifiers to all event types
	// In the future, modifiers could declare which events they care about
	eventTypes := []events.EventType{
		events.EventTypeBeforeAttackRoll,
		events.EventTypeOnAttackRoll,
		events.EventTypeAfterAttackRoll,
		events.EventTypeBeforeHit,
		events.EventTypeOnHit,
		events.EventTypeBeforeDamageRoll,
		events.EventTypeOnDamageRoll,
		events.EventTypeAfterDamageRoll,
		events.EventTypeBeforeTakeDamage,
		events.EventTypeOnTakeDamage,
		events.EventTypeAfterTakeDamage,
		events.EventTypeBeforeAbilityCheck,
		events.EventTypeOnAbilityCheck,
		events.EventTypeAfterAbilityCheck,
		events.EventTypeBeforeSavingThrow,
		events.EventTypeOnSavingThrow,
		events.EventTypeAfterSavingThrow,
	}

	for _, mod := range m.modifiers {
		if mod.IsActive(character) {
			listener := m.CreateListener(mod, character)
			subscribedTypes := make([]events.EventType, 0)

			for _, eventType := range eventTypes {
				m.eventBus.Subscribe(eventType, listener)
				subscribedTypes = append(subscribedTypes, eventType)
			}

			// Track subscriptions for this modifier
			m.subscriptions[mod.ID()] = subscribedTypes
		}
	}
}
