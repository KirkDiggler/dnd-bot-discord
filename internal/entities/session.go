package entities

import "time"

type Session struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    DraftID   string    `json:"draft_id"`
    LastToken string    `json:"last_token"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Draft     *CharacterDraft `json:"draft,omitempty"`
}
