package shared

type CharacterStatus string

const (
	CharacterStatusDraft    CharacterStatus = "draft"
	CharacterStatusActive   CharacterStatus = "active"
	CharacterStatusArchived CharacterStatus = "archived"
)
